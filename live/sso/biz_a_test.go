package sso

import (
	"bytes"
	"fmt"
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/patrickmn/go-cache"
	"net/http"
	"testing"
	"time"
)

// var ssoSessions = map[string]any{}
var aSessions = cache.New(time.Minute * 15, time.Second)
//var aSessions = ssoSessions

// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func testBizAServer(t *testing.T)  {
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

	server.Get("/token", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("token 不对")
			return
		}
		signature := Encrypt("server_a")
		// 我拿到了这个 token
		req, err := http.NewRequest(http.MethodPost,
			"http://sso.com:8000/token/validate?token=" + token , bytes.NewBuffer([]byte(signature)))
		if err != nil {
			_ = ctx.RespServerError("解析 token 失败")
			return
		}
		t.Log(req)
	})

	err := server.Start(":8081")
	t.Log(err)
}



// 登录校验的 middleware
func LoginMiddlewareServerA(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://sso.com:8000/login?client_id=server_a")
		cookie, err := ctx.Req.Cookie("token")
		if err != nil {
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := aSessions.Get(ssid)
		if !ok {
			// 你没有登录
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}
		// 这边就是登录了
		next(ctx)
	}
}
