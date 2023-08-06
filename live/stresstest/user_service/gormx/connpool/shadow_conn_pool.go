package connpool

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

// ShadowConnPool 影子库
type ShadowConnPool struct {
	liveDB *gorm.DB
}

func (s *ShadowConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {

	//TODO implement me
	panic("implement me")
}

func (s *ShadowConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShadowConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShadowConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	//TODO implement me
	panic("implement me")
}
