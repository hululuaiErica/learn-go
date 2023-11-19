package sso

import (
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
	"testing"
)

var aSessionStore = make(map[string]Session)

func TestAServer(t *testing.T) {
	server := web.NewHTTPServer(web.ServerWithMiddleware(ALoginMiddleware))
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 A 平台")
	})
	server.Start(":8081")
}

func ALoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		ck, err := ctx.Req.Cookie("ssid")
		if err != nil {
			// 这个地方你要考虑跳转，跳过去 SSO 里面
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ssid := ck.Value
		sess, ok := aSessionStore[ssid]
		if !ok {
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ctx.UserValues["sess"] = sess
		next(ctx)
	}
}

type Session struct {
	// 我 session 里面放的内容，就是 UID，你有需要你可以继续加
	Uid uint64
}
