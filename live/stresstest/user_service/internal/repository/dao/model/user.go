package model

import (
	"gorm.io/gorm"
	"log"
)

type User struct {
	Id         uint64 `gorm:"auto_increment"`
	Name       string
	Avatar     string
	Email      string
	Password   string
	Salt       string
	CreateTime uint64 // time second
	UpdateTime uint64 // time second
}

func (u *User) ShadowTableName() string {
	return "my_users_shadow"
}

func (u *User) BeforeFind(db *gorm.DB) {
	log.Println("进来了 before find")
}

func (u *User) BeforeSave(db *gorm.DB) error {
	// 我得先知道。这是不是一个压测请求
	ctx := db.Statement.Context
	stress := ctx.Value("stress-test")
	if stress == "true" {
		db.Statement.Table = "users_shadow"
	}
	return nil
}

// type UserExtend struct {
// 	Phone string
// }

// func (usr *User) ToPB() *dto.User {
// 	return &dto.User{
// 		Id: usr.Id,
// 		Name: usr.Name,
// 		Avatar: usr.Avatar,
// 		Email: usr.Email,
// 		CreateTime: usr.CreateTime,
// 	}
// }
//
// func (usr *User) ToPBWithSensitive() *dto.User {
// 	return &dto.User{
// 		Id: usr.Id,
// 		Name: usr.Name,
// 		Avatar: usr.Avatar,
// 		Email: usr.Email,
// 		CreateTime: usr.CreateTime,
// 		Password: usr.Password,
// 	}
// }
