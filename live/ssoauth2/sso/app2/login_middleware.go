package app2

import (
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
)

const ssoLoginURL = "http://sso.com:8080/check_login?source=app2"

type LoginMiddlewareBuilder struct {
	sess cache.Cache
}

func NewLoginMiddlewareBuilder(sess cache.Cache) *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{
		sess: sess,
	}
}

// Middleware 完成登录状态的校验
func (l *LoginMiddlewareBuilder) Middleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/token" {
			next(ctx)
			return
		}
		// 取凭证
		ssidCk, err := ctx.Req.Cookie("sessid")
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, ssoLoginURL, http.StatusTemporaryRedirect)
			return
		}
		// 验证凭证是有效的
		ssid := ssidCk.Value
		_, err = l.sess.Get(ctx.Req.Context(), ssid)
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, ssoLoginURL, http.StatusTemporaryRedirect)
			return
		}
		next(ctx)
	}
}
