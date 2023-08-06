package connpool

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
	"sync/atomic"
)

// ReadWriteSplitPool 影子库
type ReadWriteSplitPool struct {
	master *gorm.DB
	slaves Slaves
	idx    uint64
}

func (s *ReadWriteSplitPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	cp := s.master.ConnPool
	switch tb := cp.(type) {
	case gorm.ConnPoolBeginner:
		return tb.BeginTx(ctx, opts)
	case gorm.TxBeginner:
		tx, err := tb.BeginTx(ctx, opts)
		return &readWritePoolTransaction{
			Tx: tx,
		}, err
	default:
		return nil, gorm.ErrInvalidTransaction
	}
}

func NewReadWriteSplitPool(master *gorm.DB, slaves Slaves) *ReadWriteSplitPool {
	return &ReadWriteSplitPool{
		master: master,
		slaves: slaves,
	}
}

func (s *ReadWriteSplitPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.master.ConnPool.PrepareContext(ctx, query)
}

func (s *ReadWriteSplitPool) isMaster(ctx context.Context) bool {
	return ctx.Value("use_master") == "true"
}

func (s *ReadWriteSplitPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.master.ConnPool.ExecContext(ctx, query, args...)
}

func (s *ReadWriteSplitPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.dstForQuery(ctx).QueryContext(ctx, query, args...)
}

func (s *ReadWriteSplitPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.dstForQuery(ctx).QueryRowContext(ctx, query, args...)
}

func (s *ReadWriteSplitPool) dstForQuery(ctx context.Context) gorm.ConnPool {
	if s.isMaster(ctx) {
		return s.master.ConnPool
	}
	// 如果不是呢？
	// 挑一个从库：负载均衡
	// 扩展性不好的原因
	//idx := atomic.AddUint64(&s.idx, 1)
	//slave := s.slaves[idx%uint64(len(s.slaves))]
	//return slave.ConnPool
	return s.slaves.Next().ConnPool
}

func UseMaster(ctx context.Context) context.Context {
	return context.WithValue(ctx, "use-master", "true")
}

type readWritePoolTransaction struct {
	*sql.Tx
}

func (s *readWritePoolTransaction) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.Tx, nil
}

type Slaves interface {
	Next() *gorm.DB
}

type RoundRobinSlaves struct {
	dbs []*gorm.DB
	idx uint64
}

func (s *RoundRobinSlaves) Next() *gorm.DB {
	idx := atomic.AddUint64(&s.idx, 1)
	slave := s.dbs[idx%uint64(len(s.dbs))]
	return slave
}

type ParseDSNSlaves struct {
	dbs []*gorm.DB
}

// NewParseDSNSlaves 解析 dsn => root:root@tcp(your_company:3306)
func NewParseDSNSlaves(dsn string) *ParseDSNSlaves {
	// 拿到 your_company:3306
	// 查询 DNS，获得所有的 IP
	// return
	return &ParseDSNSlaves{}
}
