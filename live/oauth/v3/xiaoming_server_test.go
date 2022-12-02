package v3

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"gitee.com/geektime-geekbang/geektime-go/web/session"
	"gitee.com/geektime-geekbang/geektime-go/web/session/cookie"
	"gitee.com/geektime-geekbang/geektime-go/web/session/memory"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestXiaomingServer(t *testing.T) {
	server := web.NewHTTPServer()
	sessMgr := session.Manager{
		Store:      memory.NewStore(time.Minute * 15),
		Propagator: cookie.NewPropagator("sso_sess", cookie.WithCookieOption(func(c *http.Cookie) {
			c.Domain = "xiaoming.com"
		})),
		SessCtxKey: "sso_sess",
	}

	server.UseAny("/", func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// 登录界面或者登录请求
			if ctx.Req.URL.Path == "/xiaoming" {
				next(ctx)
				return
			}
			_, err := sessMgr.GetSession(ctx)
			if err != nil {
				// redirect 是待会儿在 SSO 服务器上登录成功之后要跳转回来的链接
				ctx.Redirect(fmt.Sprintf("http://sso.auth.com:%s/login?redirect=%s", ssoPort,
					url.QueryEscape("http://www.daming.com:8083/xiaoming")))
				return
			}
			next(ctx)
		}
	})

	server.Get("/xiaoming", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token").String()
		if err != nil {
			_ = ctx.RespServerError("登录失败")
			return
		}
		_, err = sessMgr.InitSession(ctx, token)
		if err != nil {
			_ = ctx.RespServerError("登录失败")
			return
		}
		_ = ctx.RespOk("hello, 你还记得大明湖畔的夏雨荷吗")
	})

	// 假如说我们登录成功之后我们就访问对应的资源
	// 这个就是模拟登录后的请求
	server.Get("/profile", func(ctx *web.Context) {
		_ = ctx.RespOk("hello, 这是夏雨荷")
	})
	server.Start(":8083")
}