package ssoauth2

import (
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"html/template"
	"log"
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
		sess := ctx.UserValues["sess"].(Session)
		log.Println(sess.Uid)
		// 在这里，你想到拿到登录的 Session, 也就是为了拿到 Uid
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
			sessionStore[ssid] = Session{
				// 随便给一个
				Uid: 123,
			}
			ctx.SetCookie(ck)
			ctx.RespString(http.StatusOK, "登录成功")
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

var sessionStore = make(map[string]Session)

// 完成登录状态的校验
func LoginMiddleware(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		ck, err := ctx.Req.Cookie("ssid")
		if err != nil {
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ssid := ck.Value
		sess, ok := sessionStore[ssid]
		if !ok {
			ctx.RespString(http.StatusUnauthorized, "请登录")
			return
		}
		ctx.UserValues["sess"] = sess
		next(ctx)
	}
}

type Session struct {
	// 我 session 里面放的内容，就是 UID，你有需要你可以继续加
	Uid uint64
}
