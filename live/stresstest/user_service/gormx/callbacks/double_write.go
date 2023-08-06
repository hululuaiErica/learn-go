package callbacks

import (
	"gorm.io/gorm"
	"log"
)

type DoubleWriteCallbackBuilder struct {
	dstDB *gorm.DB
}

func NewDoubleWriteCallbackBuilder(db *gorm.DB) *DoubleWriteCallbackBuilder {
	return &DoubleWriteCallbackBuilder{
		dstDB: db,
	}
}

func (d *DoubleWriteCallbackBuilder) Build() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		defer func() {
			if msg := recover(); msg != nil {
				// 打印一下日志
			}
		}()
		query := db.Statement.SQL.String()
		params := db.Statement.Vars

		res := d.dstDB.Exec(query, params...)
		if err := res.Error; err != nil {
			log.Println(err)
		}
		if res.RowsAffected == 0 {
			log.Println("影响 0 行")
		}
	}
}
