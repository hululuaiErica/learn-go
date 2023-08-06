package main

import (
	"context"
	"errors"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stresstest/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/grpcx/clientconn"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_web/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func main() {
	liveCC, err := grpc.Dial("localhost:8081",
		grpc.WithInsecure(),
		//grpc.WithUnaryInterceptor(func(ctx context.Context,
		//	method string, req, reply interface{},
		//	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		//	opts ...grpc.CallOption) error {
		//	stress, _ := ctx.Value("stress-test").(string)
		//	ctx = metadata.AppendToOutgoingContext(ctx, "stress-test", stress)
		//	return invoker(ctx, method, req, reply, cc, opts...)
		//})
	)

	shadowCC, err := grpc.Dial("localhost:9081",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(func(ctx context.Context,
			method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption) error {
			stress, _ := ctx.Value("stress-test").(string)
			ctx = metadata.AppendToOutgoingContext(ctx, "stress-test", stress)
			return invoker(ctx, method, req, reply, cc, opts...)
		}))

	shadow := clientconn.NewShadowClientConn(liveCC, shadowCC)

	us := userapi.NewUserServiceClient(shadow)
	userHdl := handler.NewUserHandler(us)

	r := gin.New()
	r.ContextWithFallback = true

	// 要加一个标记位提取与设置的东西
	r.Use(func(ginCtx *gin.Context) {
		stressTestHeader := ginCtx.GetHeader("x-stress-test")
		ctx := ginCtx.Request.Context()
		ctx = context.WithValue(ctx, "stress-test", stressTestHeader)
		ginCtx.Request = ginCtx.Request.WithContext(ctx)
	})

	// 登录校验
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/create" || path == "/users/login" {
			ctx.Next()
			return
		}
		sess := sessions.Default(ctx)
		if sess == nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New("请登录"))
		}
	})

	userGin := r.Group("/users")

	userGin.POST("/create", userHdl.SignUp)
	userGin.POST("/login", userHdl.Login)
	userGin.GET("/profile", userHdl.Profile)
	if err = r.Run(":8082"); err != nil {
		panic(err)
	}
}
