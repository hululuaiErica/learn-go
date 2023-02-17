package main

import (
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_web/handler"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	us := userapi.NewUserServiceClient(cc)
	userHdl := handler.NewUserHandler(us)
	r := gin.New()
	userGin := r.Group("/users")
	userGin.POST("/create", userHdl.SignUp)

	if err = r.Run(":8082"); err != nil {
		panic(err)
	}
}