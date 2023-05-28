package sso_server

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"testing"
	"time"
)

func TestSsoServer(t *testing.T) {
	tpls := template.New("sso_server")
	tpls, err := tpls.ParseGlob("./template/*")
	if err != nil {
		t.Fatal(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}

	sess := cache.NewBuildInMapCache(time.Minute * 15)
	server := web.NewHTTPServer(
		web.ServerWithTemplateEngine(engine),
		web.ServerWithMiddleware(NewSSOLoginMiddlewareBuilder(sess).Middleware))

	server.Get("/login", func(ctx *web.Context) {
		_ = ctx.Render("login.gohtml", nil)
	})

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
			_ = ctx.RespOk("登录成功")
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "登录失败")
		return
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

type Session struct {
	Uid uint64
}
