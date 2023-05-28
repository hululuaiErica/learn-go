package cache

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/orm"
)

type MiddlewareBuilder struct {
}

func (m *MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			//if qc.Type != "SELECT" {
			//	return next(ctx, qc)
			//}
			//bd := qc.Builder.(orm.Selector[User])
			//tr := bd.Table().(orm.Table)
			//
			//// 从缓存读到了
			//if readFromCache() {
			//	return &orm.QueryResult{}
			//}
			//
			return next(ctx, qc)
		}
	}
}
