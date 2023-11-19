package sso

import (
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"net/url"
	"testing"
)

var SSOSessionStore = make(map[string]any)

func TestSSOServer(t *testing.T) {
	tpls := template.New("test_server")
	tpls, err := tpls.ParseGlob("./template/*")
	if err != nil {
		t.Fatal(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}
	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))

	server.Get("/hello", func(ctx *web.Context) {
		ctx.RespString(http.StatusOK, "SSO 启动成功了")
	})

	server.Post("/login", func(ctx *web.Context) {
		email, _ := ctx.FormValue("email")
		pwd, _ := ctx.FormValue("password")
		if email == "123@qq.com" && pwd == "123456" {
			ssid := uuid.New().String()
			// 这边要怎么办？
			// 在这边你要设置好 session
			ck := &http.Cookie{
				Name:   "ssid",
				Value:  ssid,
				MaxAge: 1800,
				// 在 https 里面才能用这个 cookie
				//Secure: true,
				// 前端没有办法通过 JS 来访问 cookie
				HttpOnly: true,
			}
			SSOSessionStore[ssid] = Session{
				// 随便给一个
				Uid: 123,
			}
			ctx.SetCookie(ck)
			// 这个地方怎么办？是不是要跳回去？
			path, _ := ctx.FormValue("redirect_uri")
			ctx.Redirect(path)
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "登录失败")
		return
	})

	// 这个地方怎么写？
	// 要有一个新的 HTTP 接口
	// 要判断登录态，如果没登录就返回登录页面，
	// 如果登录了，就跳转回去 A/B
	server.Any("/check_login", func(ctx *web.Context) {
		// 21:12 分
		ck, err := ctx.Req.Cookie("ssid")
		path, _ := ctx.QueryValue("redirect_uri")
		path, _ = url.PathUnescape(path)
		if err != nil {
			_ = ctx.Render("login.gohtml", map[string]string{
				"RedirectURI": path,
			})
			return
		}
		ssid := ck.Value
		_, ok := SSOSessionStore[ssid]
		if !ok {
			_ = ctx.Render("login.gohtml", map[string]string{
				"RedirectURI": path,
			})
			return
		}

		// 这边就是登录了
		// 我要跳回去
		ctx.Redirect(path)
	})
	server.Start(":8083")
}
