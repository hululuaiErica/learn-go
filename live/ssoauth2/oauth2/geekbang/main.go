package main

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"html/template"
	"io"
	"net/http"
)

func main() {

	tpls := template.New("geekbang")
	tpls, err := tpls.ParseGlob("./template/*")
	if err != nil {
		panic(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}
	server := web.NewHTTPServer(
		web.ServerWithTemplateEngine(engine))

	server.Get("/login", func(ctx *web.Context) {
		_ = ctx.Render("login_page.gohtml", nil)
	})

	server.Get("/wechat_login", func(ctx *web.Context) {
		ctx.Redirect("http://localhost:8081/login?appid=geekbang")
	})

	server.Get("/authed", func(ctx *web.Context) {
		// 这个就是临时授权码
		code, _ := ctx.QueryValue("code")

		resp, err := http.Get(fmt.Sprintf("http://localhost:8081/access_token?code=%s&appid=geekbang", code))
		// 不知道除了什么问题
		if err != nil {
			ctx.RespString(http.StatusInternalServerError, "服务器故障")
			return
		}

		accessToken, _ := io.ReadAll(resp.Body)
		if len(accessToken) > 0 {
			// 拿到了 access token
			// 可以去访问资源了，就是个人信息

			ctx.RespString(http.StatusOK, "你已经得到授权了")
		}
	})

	if err := server.Start(":8080"); err != nil {
		panic(err)
	}
}
