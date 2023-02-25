package main

import (
	"context"
	"errors"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_web/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	cc, err := NewClientConnWrapper("localhost:8081", "localhost:9081")
	if err != nil {
		panic(err)
	}
	us := userapi.NewUserServiceClient(cc)
	userHdl := handler.NewUserHandler(us)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(func(ctx *gin.Context) {
		// ctx 里面压测标记位
		// ctx.Request.Header => ctx.Request.Context()
		if ctx.Request.Header.Get("x_stress_test") == "true" {
			cctx := context.WithValue(ctx.Request.Context(), "stress_test", "true")
			ctx.Request = ctx.Request.WithContext(cctx)
		}
	})
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