package repository

import (
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = errors.New("未找到指定的用户")
)
