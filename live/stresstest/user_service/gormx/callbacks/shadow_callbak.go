package callbacks

import "gorm.io/gorm"

type ShadowTableCallbackBuilder struct {
	shadowTableName map[string]string
}

func NewShadowTableCallbackBuilder() *ShadowTableCallbackBuilder {
	return &ShadowTableCallbackBuilder{
		//shadowTableName: map[string]string{},
	}
}

func (b *ShadowTableCallbackBuilder) Set(table string, shadowTable string) *ShadowTableCallbackBuilder {
	if b.shadowTableName == nil {
		b.shadowTableName = make(map[string]string, 4)
	}
	b.shadowTableName[table] = shadowTable
	return b
}

func (b *ShadowTableCallbackBuilder) Build() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		stress := ctx.Value("stress-test")
		// my_users_shadow
		if stress == "true" {
			if tblName, ok := db.Statement.Model.(interface {
				ShadowTableName() string
			}); ok {
				db.Statement.Table = tblName.ShadowTableName()
				return
			}

			tblName, ok := b.shadowTableName[db.Statement.Table]
			if ok {
				db.Statement.Table = tblName
			} else {
				db.Statement.Table = db.Statement.Table + "_shadow"
			}
		}
	}
}
