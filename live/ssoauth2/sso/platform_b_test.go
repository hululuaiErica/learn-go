package sso

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
	"net/url"
	"testing"
)

// var bSessionStore = make(map[string]Session)
var bSessionStore = SSOSessionStore

func TestBServer(t *testing.T) {
	server := web.NewHTTPServer(web.ServerWithMiddleware(BLoginMiddleware))
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 B 平台")
	})
	server.Get("/health", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 B 平台，你跳回来了")
	})
	server.Start(":8082")
}

func BLoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" || ctx.Req.URL.Path == "/health" {
			next(ctx)
			return
		}
		ck, err := ctx.Req.Cookie("ssid")
		path := "http://localhost:8082/health"
		const pattern = "http://localhost:8083/check_login?redirect_uri=%s"
		// URL 编码
		path = fmt.Sprintf(pattern,
			url.PathEscape(path))
		if err != nil {
			// 这个地方你要考虑跳转，跳过去 SSO 里面
			ctx.Redirect(path)
			//ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ssid := ck.Value
		sess, ok := bSessionStore[ssid]
		if !ok {
			ctx.Redirect(path)
			//ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ctx.UserValues["sess"] = sess
		next(ctx)
	}
}
