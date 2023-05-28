package sso_server

import (
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
	"testing"
	"time"
)

func TestSsoServer(t *testing.T) {
	sess := cache.NewBuildInMapCache(time.Minute * 15)
	server := web.NewHTTPServer(web.ServerWithMiddleware(NewSSOLoginMiddlewareBuilder(sess).Middleware))

	server.Get("/login", func(ctx *web.Context) {

	})

	if err := server.Start(":8080"); err != nil {
		panic(err)
	}
}

type SSOLoginMiddlewareBuilder struct {
	sess cache.Cache
}

func NewSSOLoginMiddlewareBuilder(sess cache.Cache) *SSOLoginMiddlewareBuilder {
	return &SSOLoginMiddlewareBuilder{
		sess: sess,
	}
}

// Middleware 完成登录状态的校验
func (l *SSOLoginMiddlewareBuilder) Middleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		// 取凭证
		ssidCk, err := ctx.Req.Cookie("sessid")
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, "/login", http.StatusTemporaryRedirect)
			return
		}
		// 验证凭证是有效的
		ssid := ssidCk.Value
		_, err = l.sess.Get(ctx.Req.Context(), ssid)
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, "/login", http.StatusTemporaryRedirect)
			return
		}
		next(ctx)
	}
}
