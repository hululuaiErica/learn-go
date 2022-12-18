package sso

import (
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"html/template"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestSSOServer(t *testing.T) {
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}
	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))
	server.Get("/login", func(ctx *web.Context) {
		redirect, _ := ctx.QueryValue("redirect")
		_ = ctx.Render("login.gohtml", map[string]string{"Redirect": redirect})
	})
	server.Post("/login", func(ctx *web.Context) {
		// 我在这儿模拟登录
		if err != nil {
			ctx.RespServerError("系统错误")
		}
		// 校验账号和密码
		email, _ := ctx.FormValue("email")
		password, _ := ctx.FormValue("password")
		redirect, _ := ctx.FormValue("redirect")
		redirect, _ = url.QueryUnescape(redirect)
		if email == "abc@biz.com" && password == "123" {
			// 认为登录成功
			// 要防止 token 被盗走，不能使用 uuid
			id := uuid.New().String()
			http.SetCookie(ctx.Resp, &http.Cookie{
				Name: "token",
				Value: id,
				Expires: time.Now().Add(time.Minute * 15),
			})
			aSessions.Set(id, &User{Name: "Tom"}, time.Minute * 15)
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}
		ctx.RespServerError("用户账号名密码不对")
	})
	// 要在这里提供登录的地方
	server.Start(":8000")
}
