package service

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/domainobject/entity"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

type userService struct {
	repo repository.UserRepository
	userapi.UnimplementedUserServiceServer
}

func NewUserService(repo repository.UserRepository) userapi.UserServiceServer {
	return &userService{
		repo: repo,
	}
}


func (u *userService) EditProfile(ctx context.Context, user entity.User) error {
	return u.repo.UpdateUser(ctx, user)
}

func (u *userService) FindById(ctx context.Context, id uint64)(entity.User, error) {
	return u.repo.GetUserById(ctx, id)
}

func (u *userService) Login(ctx context.Context, input entity.User) (entity.User, error) {
	usr, err := u.repo.GetUserByEmail(ctx, input.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return entity.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return entity.User{}, err
	}

	encryptedPwd := u.encryptPwdByPbkdf2(input.Password, usr.Salt)
	if encryptedPwd != usr.Password {
		return entity.User{}, ErrInvalidUserOrPassword
	}
	return usr, nil
}

func(u *userService) CreateUser(ctx context.Context, req *userapi.CreateUserReq) (*userapi.CreateUserResp, error) {
	user := entity.User{
		Name: req.User.Name,
		Avatar: req.User.Avatar,
		Email: req.User.Email,
		Password: req.User.Password,
	}
	err := user.Check()
	if err != nil {
		return nil, fmt.Errorf("%w, 原因 %v", ErrInvalidNewUser, err)
	}
	// 开始加密
	// 每一个用户都是一个单独的 salt，会更加安全
	// 你也可以考虑组合多个 HASH 加密算法来存储
	salt := uuid.New().String()
	user.Salt = salt
	user.Password = u.encryptPwdByPbkdf2(user.Password, salt)
	user, err =  u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	req.User.Id = user.Id
	return &userapi.CreateUserResp{
		User: req.User,
	}, err
}

func (u *userService) encryptPwdByPbkdf2(raw string, salt string) string {
	// pbkdf2 需要比较多的 CPU 的资源。不过考虑到注册用户整体上是非常非常低频的，那么你也不会介意使用这种复杂的加密算法
	return fmt.Sprintf("%X", pbkdf2.Key([]byte(raw), []byte(salt), 4096, 32, sha1.New))
}

func (u *userService) ServiceName() string {
	return "user"
}

