package connpool

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

// ShadowConnPool 影子库
type ShadowConnPool struct {
	liveDB   gorm.ConnPool
	shadowDB gorm.ConnPool
}

func (s *ShadowConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	cp := s.liveDB
	if s.isShadow(ctx) {
		cp = s.shadowDB
	}
	switch tb := cp.(type) {
	case gorm.ConnPoolBeginner:
		return tb.BeginTx(ctx, opts)
	case gorm.TxBeginner:
		tx, err := tb.BeginTx(ctx, opts)
		return &shadowPoolTransaction{
			Tx: tx,
		}, err
	default:
		return nil, gorm.ErrInvalidTransaction
	}
}

//func (s *ShadowConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
//	cp := s.master
//	if s.isShadow(ctx) {
//		cp = s.shadowDB
//	}
//	tb, ok := cp.(gorm.TxBeginner)
//	if !ok {
//		return nil, gorm.ErrInvalidTransaction
//	}
//	return tb.BeginTx(ctx, opts)
//}

func NewShadowConnPool(liveDB gorm.ConnPool, shadowDB gorm.ConnPool) *ShadowConnPool {
	return &ShadowConnPool{
		liveDB:   liveDB,
		shadowDB: shadowDB,
	}
}

func (s *ShadowConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if s.isShadow(ctx) {
		return s.shadowDB.PrepareContext(ctx, query)
	}
	return s.liveDB.PrepareContext(ctx, query)
}

func (s *ShadowConnPool) isShadow(ctx context.Context) bool {
	return ctx.Value("stress-test") == "true"
}

func (s *ShadowConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if s.isShadow(ctx) {
		return s.shadowDB.ExecContext(ctx, query, args...)
	}
	return s.liveDB.ExecContext(ctx, query, args...)
}

func (s *ShadowConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if s.isShadow(ctx) {
		return s.shadowDB.QueryContext(ctx, query, args...)
	}
	return s.liveDB.QueryContext(ctx, query, args...)
}

func (s *ShadowConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if s.isShadow(ctx) {
		return s.shadowDB.QueryRowContext(ctx, query, args...)
	}
	return s.liveDB.QueryRowContext(ctx, query, args...)
}

type shadowPoolTransaction struct {
	*sql.Tx
}

func (s shadowPoolTransaction) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.Tx, nil
}
