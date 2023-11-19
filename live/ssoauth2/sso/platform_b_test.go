package sso

import (
	"gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
	"testing"
)

var bSessionStore = make(map[string]Session)

func TestBServer(t *testing.T) {
	server := web.NewHTTPServer(web.ServerWithMiddleware(BLoginMiddleware))
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 B 平台")
	})
	server.Start(":8082")
}

func BLoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		ck, err := ctx.Req.Cookie("ssid")
		if err != nil {
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ssid := ck.Value
		sess, ok := bSessionStore[ssid]
		if !ok {
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ctx.UserValues["sess"] = sess
		next(ctx)
	}
}
