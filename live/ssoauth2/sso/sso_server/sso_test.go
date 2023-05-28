package sso_server

import (
	"context"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"testing"
	"time"
)

var bizRedirectUrl = map[string]string{
	"app1": "http://app1.com:8081/token?token=%s",
	"app2": "http://app2.com:8082/token?token=%s",
}

func TestSsoServer(t *testing.T) {
	tpls := template.New("sso_server")
	tpls, err := tpls.ParseGlob("./template/*")
	if err != nil {
		t.Fatal(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}

	tokens := cache.NewBuildInMapCache(time.Minute * 3)

	sess := cache.NewBuildInMapCache(time.Minute * 15)
	server := web.NewHTTPServer(
		web.ServerWithTemplateEngine(engine),
		web.ServerWithMiddleware(NewSSOLoginMiddlewareBuilder(sess).Middleware))

	server.Get("/login", func(ctx *web.Context) {
		source, _ := ctx.QueryValue("source")
		_ = ctx.Render("login.gohtml", map[string]string{
			"Source": source,
		})
	})

	server.Get("/check_login", func(ctx *web.Context) {
		// 能进来这里，就说明已经登录了
		source, _ := ctx.QueryValue("source")
		url := bizRedirectUrl[source]
		token := uuid.New().String()
		_ = tokens.Set(context.Background(), token, source, time.Minute)
		ctx.Redirect(fmt.Sprintf(url, token))
	})

	// 这个接口要加限流，比如针对 IP 的限流
	server.Post("/login", func(ctx *web.Context) {
		// 处理登录请求
		email, _ := ctx.FormValue("email")
		pwd, _ := ctx.FormValue("password")
		if email == "123@qq.com" && pwd == "123456" {
			ssid := uuid.New().String()
			sess.Set(context.Background(), ssid, Session{Uid: 123}, time.Minute*15)
			ctx.SetCookie(&http.Cookie{
				Name:   "sessid",
				Value:  ssid,
				Domain: "sso.com",
				//Expires: time.Now().Add(time.Minute * 10),
			})
			source, _ := ctx.QueryValue("source")
			url := bizRedirectUrl[source]
			token := uuid.New().String()
			_ = tokens.Set(context.Background(), token, source, time.Minute)
			ctx.Redirect(fmt.Sprintf(url, token))
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "登录失败")
		return
	})

	// token 校验，保护好
	// 请求来源可以要求一个 app 一个 IP
	server.Post("/token/validate", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		// 可能会有一个解密的过程

		_, er := tokens.Get(ctx.Req.Context(), token)
		// 稍微比较一下就可以
		if er != nil {
			ctx.RespString(http.StatusForbidden, "没有权限")
			return
		}
		// 带上用户信息，比如说 uid
		ctx.RespString(http.StatusOK, "123")
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
		if ctx.Req.URL.Path == "/login" || ctx.Req.URL.Path == "/token/validate" {
			next(ctx)
			return
		}

		// 取凭证
		ssidCk, err := ctx.Req.Cookie("sessid")
		source, _ := ctx.QueryValue("source")
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, "/login?source="+source, http.StatusTemporaryRedirect)
			return
		}
		// 验证凭证是有效的
		ssid := ssidCk.Value
		_, err = l.sess.Get(ctx.Req.Context(), ssid)
		if err != nil {
			// 重定向过去登录那里
			http.Redirect(ctx.Resp, ctx.Req, "/login?source="+source, http.StatusTemporaryRedirect)
			return
		}

		next(ctx)
	}
}

type Session struct {
	Uid uint64
}
