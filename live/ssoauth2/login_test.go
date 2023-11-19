package ssoauth2

import (
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"html/template"
	"net/http"
	"testing"
)

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func TestBizServer(t *testing.T) {
	tpls := template.New("test_server")
	tpls, err := tpls.ParseGlob("./template/*")
	if err != nil {
		t.Fatal(err)
	}
	engine := &web.GoTemplateEngine{
		T: tpls,
	}

	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine),
		web.ServerWithMiddleware(LoginMiddleware))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/profile", func(ctx *web.Context) {
		_ = ctx.RespJSONOK(&User{
			Name: "Tom",
			Age:  18,
		})
	})

	server.Get("/login", func(ctx *web.Context) {
		_ = ctx.Render("login.gohtml", nil)
	})

	server.Post("/login", func(ctx *web.Context) {
		email, _ := ctx.FormValue("email")
		pwd, _ := ctx.FormValue("password")
		if email == "123@qq.com" && pwd == "123456" {
			// 这边要怎么办？
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "登录失败")
		return
	})

	err = server.Start(":8081")
	t.Log(err)

}

type User struct {
	Name     string
	Password string
	Age      int
}

// 完成登录状态的校验
func LoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {

	}
}

type Session struct {
	Uid uint64
}
