package main

import (
	"context"
	"database/sql"
	"strings"
)

// 利用类似思路来实现主从分离
type MasterSlaveConnPool struct {
	master *sql.DB
	slaves []*sql.DB
}

func NewMasterSlaveConnPool(masterDSN string, slavesDSN string) *MasterSlaveConnPool{
	return &MasterSlaveConnPool{

	}
}

// user.master.mycompany.com:3306
// user.slave.mycompany.com:3306/user_db => 会有好几个从库
func (m *MasterSlaveConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if strings.Contains(query, "SELECT") {
		return m.slaves[0].PrepareContext(ctx, query)
	}
	return m.master.PrepareContext(ctx, query)
}

func (m *MasterSlaveConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// TODO implement me
	panic("implement me")
}

// 没有解决强制走主库的问题
func (m *MasterSlaveConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if ctx.Value("use_master") == "true" {
		return m.master.QueryContext(ctx, query, args...)
	}
	return m.slaves[0].QueryContext(ctx, query, args...)
}

func (m *MasterSlaveConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// TODO implement me
	panic("implement me")
}

func CtxWithMaster(ctx context.Context) context.Context {
	return context.WithValue(ctx, "use_master", "true")
}

