package main

import (
	"errors"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_web/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net/http"
)

func main() {
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	us := userapi.NewUserServiceClient(cc)
	shadowCc, err := grpc.Dial("localhost:9081", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	shadowUs := userapi.NewUserServiceClient(shadowCc)

	userHdl := handler.NewUserHandler(&UserServiceClient{
		client: us,
		shadowClient: shadowUs,
	})
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/create" || path == "/user/login" {
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