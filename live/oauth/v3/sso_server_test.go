package v3

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"gitee.com/geektime-geekbang/geektime-go/web/session"
	"gitee.com/geektime-geekbang/geektime-go/web/session/cookie"
	"gitee.com/geektime-geekbang/geektime-go/web/session/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"html/template"
	"net/http"
	"testing"
	"time"
)

const ssoPort = "8081"

func TestSSOServer(t *testing.T) {
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}
	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))
	sessMgr := session.Manager{
		Store:      memory.NewStore(time.Minute * 15),
		Propagator: cookie.NewPropagator("sso_sess", cookie.WithCookieOption(func(c *http.Cookie) {
			c.Domain = "sso.com"
		})),
		SessCtxKey: "sso_sess",
	}

	server.UseAny("/", func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// 登录界面或者登录请求
			if ctx.Req.URL.Path == "/login" {
				next(ctx)
				return
			}
			_, err := sessMgr.GetSession(ctx)
			if err != nil {
				ctx.Redirect("login")
				return
			}
			next(ctx)
		}
	})

	server.Get("/login", func(ctx *web.Context) {
		redirect, err := ctx.QueryValue("redirect").String()
		if err != nil {
			 _ = ctx.RespString(http.StatusBadRequest, "请求参数不对")
			return
		}
		_ = ctx.Render("login.gohtml", redirect)
	})

	server.Post("/login", func(ctx *web.Context) {
		email, err := ctx.FormValue("email").String()
		if err != nil {
			_ = ctx.RespString(http.StatusBadRequest, "邮箱输入错误")
			return
		}
		password, err := ctx.FormValue("password").String()
		if err != nil {
			_ = ctx.RespString(http.StatusBadRequest, "密码错误")
			return
		}
		if password == "123" && email == "123@demo.com" {
			ssid := uuid.New().String()
			_, err = sessMgr.InitSession(ctx, ssid)
			if err != nil {
				_ = ctx.RespServerError("登录失败")
				return
			}
			redirect, err := ctx.QueryValue("redirect").String()
			if err != nil {
				_ = ctx.RespServerError("登录失败")
				return
			}
			token := uuid.New().String()
			ctx.Redirect(fmt.Sprintf("%s?token=%s", redirect, token))
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "用户名密码不对")
	})

	server.Start(":"+ssoPort)
}