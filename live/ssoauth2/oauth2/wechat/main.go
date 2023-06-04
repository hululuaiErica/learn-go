package main

import (
	"context"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"time"
)

var bizRedirectUrl = map[string]string{
	"app1":     "http://app1.com:8081/token?token=%s",
	"geekbang": "http://localhost:8080/authed?code=%s",
	"app2":     "http://app2.com:8082/token?token=%s",
}

func main() {

	tpls := template.New("wechat")
	tpls, err := tpls.ParseGlob("./template/*")

	// 临时授权码存储
	tmpCodes := cache.NewBuildInMapCache(time.Minute * 3)

	if err != nil {
		panic(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}
	server := web.NewHTTPServer(
		web.ServerWithTemplateEngine(engine))

	server.Get("/login", func(ctx *web.Context) {
		appID, _ := ctx.QueryValue("appid")
		_ = ctx.Render("login.gohtml", map[string]string{
			"AppID": appID,
		})
	})

	// 这个接口要加限流，比如针对 IP 的限流
	server.Post("/login", func(ctx *web.Context) {
		appid, _ := ctx.QueryValue("appid")
		url := bizRedirectUrl[appid]
		code := uuid.New().String()
		_ = tmpCodes.Set(context.Background(), code, appid, time.Minute)
		ctx.Redirect(fmt.Sprintf(url, code))
		return
	})

	server.Get("/access_token", func(ctx *web.Context) {
		appID, _ := ctx.QueryValue("appid")
		code, _ := ctx.QueryValue("code")
		val, _ := tmpCodes.Get(ctx.Req.Context(), code)
		if appID == val {
			accessToken := uuid.New().String()
			_ = ctx.RespString(http.StatusOK, accessToken)
		} else {
			_ = ctx.RespString(http.StatusInternalServerError, "Error")
		}
	})
	if err := server.Start(":8081"); err != nil {
		panic(err)
	}
}
