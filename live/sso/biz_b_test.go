package sso

import (
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"net/http"
	"testing"
	"time"
)

// var sessions = map[string]any{}
var bSessions = cache.New(time.Minute * 15, time.Second)
// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func TestBizBServer(t *testing.T)  {
	server := web.NewHTTPServer(web.ServerWithMiddleware(LoginMiddlewareServerB))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespJSONOK(&User{
			Name: "Tom",
			Age: 18,
		})
	})

	server.Post("/login", func(ctx *web.Context) {
		// 我在这儿模拟登录
		var u User
		err := ctx.BindJSON(&u)
		if err != nil {
			ctx.RespServerError("系统错误")
		}
		// 校验账号和密码
		if u.Name == "abc" && u.Password == "123" {
			// 认为登录成功
			// 要防止 token 被盗走，不能使用 uuid
			id := uuid.New().String()
			http.SetCookie(ctx.Resp, &http.Cookie{
				Name: "token",
				Value: id,
				Expires: time.Now().Add(time.Minute * 15),
			})
			bSessions.Set(id, &User{Name: "Tom"}, time.Minute * 15)
			ctx.RespJSONOK(&User{Name: "Tom"})
			return
		}
		ctx.RespServerError("用户账号名密码不对")
	})

	err := server.Start(":8082")
	t.Log(err)
}



func LoginMiddlewareServerB(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/login" {
			next(ctx)
			return
		}
		// ssid，即 session id
		//
		cookie, err := ctx.Req.Cookie("token")
		if err != nil {
			ctx.RespServerError("你没有登录-token")
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := sessions.Get(ssid)
		if !ok {
			// 你没有登录
			ctx.RespServerError("你没有登录-sess id")
			return
		}
		// 这边就是登录了
		next(ctx)
	}
}