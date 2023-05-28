package app1

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestApp1Server(t *testing.T) {
	sess := cache.NewBuildInMapCache(time.Minute * 15)
	server := web.NewHTTPServer(web.ServerWithMiddleware(NewLoginMiddlewareBuilder(sess).Middleware))

	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 APP1，你进来啦")
	})

	server.Get("/token", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		// 要去解析 token
		// 怎么发起调用

		resp, err := http.Post("http://sso.com:8080/token/validate?token="+token,
			"application/json", nil)
		// 不知道除了什么问题
		if err != nil {
			ctx.RespString(http.StatusInternalServerError, "服务器故障")
			return
		}

		body, _ := io.ReadAll(resp.Body)
		// 获得 token 解析结果
		// 假设 123 是token
		if string(body) != "123" {
			ctx.RespString(http.StatusForbidden, "非法访问")
			return
		}

		// 种下 session 和 cookie
		ssid := uuid.New().String()
		sess.Set(context.Background(), ssid, Session{Uid: 123}, time.Minute*15)
		ctx.SetCookie(&http.Cookie{
			Name:   "sessid",
			Value:  ssid,
			Domain: "app1.com",
			//Expires: time.Now().Add(time.Minute * 10),
		})
		ctx.RespString(http.StatusOK, "彻底完成登录")
	})

	if err := server.Start(":8081"); err != nil {
		panic(err)
	}
}

type Session struct {
	Uid uint64
}
