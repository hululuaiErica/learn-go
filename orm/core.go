package orm

import (
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/valuer"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
)

type core struct {
	model *model.Model
	dialect Dialect
	creator valuer.Creator
	r model.Registry
	mdls []Middleware
}
