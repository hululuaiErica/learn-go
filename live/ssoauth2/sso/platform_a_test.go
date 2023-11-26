package sso

import (
	"fmt"
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"net/http"
	"net/url"
	"testing"
)

// var aSessionStore = make(map[string]Session)
// 这个地方模拟，大家共享一个 Redis Session 之类的东西
var aSessionStore = SSOSessionStore

func TestAServer(t *testing.T) {
	server := web.NewHTTPServer(web.ServerWithMiddleware(ALoginMiddleware))
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "这是 A 平台")
	})
	// 就是处理从 SSO 跳回来的逻辑，也就是说，我要在这里设置登录态
	// 我可以直接设置吗？
	server.Get("/health", func(ctx *web.Context) {
		// 我自己设置一个登录态
		// 第一个问题：你怎么知道，这个地方就是从 SSO 过来的？
		// 解析 token
		// 调用 SSO 的另外一个接口，去解析 token
		ctx.RespString(http.StatusOK, "这是 A 平台，你跳回来了")
	})
	server.Start(":8081")
}

func ALoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" || ctx.Req.URL.Path == "/health" {
			next(ctx)
			return
		}
		ck, err := ctx.Req.Cookie("ssid")
		path := "http://localhost:8081/health"
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
		sess, ok := aSessionStore[ssid]
		if !ok {
			ctx.Redirect(path)
			//ctx.RespString(http.StatusUnauthorized, "请登录")
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
