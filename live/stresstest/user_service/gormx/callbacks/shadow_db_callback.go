package callbacks

import (
	"gitee.com/geektime-geekbang/geektime-go/live/stresstest/user_service/internal/domainobject/entity"
	"gorm.io/gorm"
)

type ShadowDBCallbackBuilder struct {
	shadowPool gorm.ConnPool
}

func (b *ShadowDBCallbackBuilder) Build() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement.Context.Value("stress-test") == "true" {
			db.Statement.ConnPool = b.shadowPool
			db.Statement.Model.(*entity.User).Shadow = true
		}
		// 你还可以结合前面的 shadow table 一起处理掉。
	}
}
