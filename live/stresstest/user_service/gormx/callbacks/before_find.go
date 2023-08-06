package callbacks

import "gorm.io/gorm"

type BeforeFindBuilder struct {
}

func (b *BeforeFindBuilder) Build() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if bf, ok := db.Statement.Model.(interface {
			BeforeFind(db *gorm.DB)
		}); ok {
			bf.BeforeFind(db)
		}
	}
}
