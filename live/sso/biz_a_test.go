package sso

import (
	"fmt"
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/patrickmn/go-cache"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// var sessions = map[string]any{}
var aSessions = cache.New(time.Minute * 15, time.Second)
// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func TestBizAServer(t *testing.T)  {
	server := web.NewHTTPServer(web.ServerWithMiddleware(LoginMiddlewareServerA))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespJSONOK(&User{
			Name: "Tom",
			Age: 18,
		})
	})

	err := server.Start(":8081")
	t.Log(err)
}



func LoginMiddlewareServerA(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://sso.biz.com:8000/login?redirect=%s",
			url.QueryEscape("http://a.biz.com:8081/profile"))
		cookie, err := ctx.Req.Cookie("token")
		if err != nil {
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := sessions.Get(ssid)
		if !ok {
			// 你没有登录
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}
		// 这边就是登录了
		next(ctx)
	}
}