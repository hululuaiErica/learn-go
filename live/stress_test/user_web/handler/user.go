package handler

import (
	"errors"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

const (
	userIdKey = "user_id"
)

type UserHandler struct {
	service userapi.UserServiceClient
}

func NewUserHandler(us userapi.UserServiceClient) *UserHandler {
	return &UserHandler{
		service: us,
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	req := loginReq{}
	err := ctx.BindJSON(&req)
	if err != nil {
		zap.L().Error("handler: 解析 JSON 数据格式失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, Resp{
			Msg: "解析请求失败",
		})
		return
	}
	usr, err := h.service.Login(ctx.Request.Context(), &userapi.LoginReq{
		Email: req.Email,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("登录失败，系统异常", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, Resp{
			Msg: "系统异常",
		})
		return
	}
	sess := sessions.Default(ctx)
	sess.Set(userIdKey, usr.User.Id)
	err = sess.Save()
	if err != nil {
		zap.L().Error("登录失败，设置 session 失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, Resp{
			Msg: "系统异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, Resp{
		Msg: "登录成功",
	})
}

//func (h *UserHandler) Update(ctx *web.Context) {
//	u := User{}
//	err := ctx.BindJSON(&u)
//	if err != nil{
//		zap.L().Error("web: 解析 JSON 数据错误", zap.Error(err))
//		_ = ctx.RespJSON(http.StatusInternalServerError, Resp{
//			Msg: "系统异常",
//		})
//	}
//
//	uid, err := h.getId(ctx)
//	if err != nil {
//		zap.L().Error("handler: 无法获得 user id", zap.Error(err))
//		_ = ctx.RespJSON(http.StatusInternalServerError, Resp{
//			Msg: "系统异常",
//		})
//		return
//	}
//
//	err = h.service.EditProfile(ctx.Req.Context(), entity.User{
//		// 一般是前端传了什么，这边就往下传什么
//		Id: uid,
//		Name: u.Name,
//		Email: u.Email,
//	})
//	if err != nil {
//		zap.L().Error("handler: 无法更新用户详情", zap.Error(err))
//		_ = ctx.RespJSON(http.StatusInternalServerError, Resp{
//			Msg: "系统异常",
//		})
//		return
//	}
//
//	// 可以考虑忽略，不过不嫌麻烦还是要和其它方法一样处理一下
//	_ = ctx.RespJSON(http.StatusOK, Resp{
//		Msg: "ok",
//	})
//}

func (h *UserHandler) Profile(ctx *gin.Context) {
	uid, err := h.getId(ctx)
	if err != nil {
		zap.L().Error("handler: 无法获得 user id", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, Resp{
			Msg: "系统异常",
		})
		return
	}
	usrResp, err := h.service.FindById(ctx.Request.Context(), &userapi.FindByIdReq{
		Id: uid,
	})
	if err != nil {
		zap.L().Error("web: 查找用户失败", zap.Error(err))
		ctx.String(http.StatusInternalServerError, "system error")
		return
	}
	usr := usrResp.User
	ctx.JSON(http.StatusOK, Resp{
		Data: User{
			Email: usr.Email,
			Name: usr.Name,
			Avatar: usr.Avatar,
		},
	})
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	u := &signUpReq{}
	err := ctx.BindJSON(u)
	if err != nil {
		zap.L().Error("web: 解析 JSON 数据格式失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, Resp{
			Msg: "解析请求失败",
		})
		return
	}

	_, err = h.service.CreateUser(ctx.Request.Context(), &userapi.CreateUserReq{
		User: &userapi.User{
			Email: u.Email,
			Password: u.Password,
		},
	})
	if err != nil {
		zap.L().Error("创建用户失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, &Resp{
			Msg: "创建用户失败",
		})
		return
	}
	ctx.String(http.StatusOK, "创建成功")
}

func (h *UserHandler) getId(ctx *gin.Context) (uint64, error){
	s := sessions.Default(ctx)
	if s == nil {
		return 0, errors.New("尚未登录")
	}
	uid := s.Get(userIdKey).(uint64)
	return uid, nil
}
