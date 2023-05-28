package app1

import (
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
)

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
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		// 取凭证
		ssidCk, err := ctx.Req.Cookie("sessid")
		if err != nil {
			ctx.RespString(http.StatusForbidden, "没有登录，找不到 cookie")
			return
		}
		// 验证凭证是有效的
		ssid := ssidCk.Value
		_, err = l.sess.Get(ctx.Req.Context(), ssid)
		if err != nil {
			ctx.RespString(http.StatusForbidden, "没有登录，找不到 session")
			return
		}
		next(ctx)
	}
}
