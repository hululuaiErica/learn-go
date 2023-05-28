package sso

import (
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"html/template"
	"net/http"
	"testing"
	"time"
)

var mySessions = cache.New(time.Minute*15, time.Second)

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
			ssid := uuid.New().String()
			mySessions.Set(ssid, Session{Uid: 123}, time.Minute*15)
			ctx.SetCookie(&http.Cookie{
				Name:   "sessid",
				Value:  ssid,
				Domain: "mycompany.com",
				//Expires: time.Now().Add(time.Minute * 10),
			})
			_ = ctx.RespOk("登录成功")
			return
		}
		_ = ctx.RespString(http.StatusBadRequest, "登录失败")
		return
	})

	go func() {
		server2 := web.NewHTTPServer(web.ServerWithTemplateEngine(engine),
			web.ServerWithMiddleware(LoginMiddleware))
		server2.Get("/profile", func(ctx *web.Context) {
			_ = ctx.RespJSONOK(&User{
				Name: "Tom",
				Age:  18,
			})
		})
		er := server2.Start(":8082")
		t.Log(er)
	}()

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
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		// 取凭证
		ssidCk, err := ctx.Req.Cookie("sessid")
		if err != nil {
			ctx.RespString(http.StatusForbidden, "没有登录，找不到 cookie")
			return
		}
		// 验证凭证是有效的
		ssid := ssidCk.Value
		_, ok := mySessions.Get(ssid)
		if !ok {
			ctx.RespString(http.StatusForbidden, "没有登录，找不到 session")
			return
		}
		next(ctx)
	}
}

type Session struct {
	Uid uint64
}
