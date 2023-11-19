package sso

import (
	"gitee.com/geektime-geekbang/geektime-go/web"
	"html/template"
	"net/http"
	"testing"
)

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
	// 这个地方怎么写？
	// 要有一个新的 HTTP 接口
	// 要判断登录态，如果没登录就返回登录页面，
	// 如果登录了，就跳转回去 A/B
	server.Get("/check_login", func(ctx *web.Context) {

	})
	server.Start(":8083")
}
