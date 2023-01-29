package orm

import (
	"errors"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"strings"
)

type ShardingSelector[T any] struct {
	builder
	table *T
	where []Predicate
	columns []Selectable

	sess Session
}

//type ShardingFunc[T ShardingKey] func(skVal T) (string, string)
//
//type ShardingKey interface {
//	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8| ~uint16| ~uint32 | ~uint64
//}

// 如果 T 是 User，那么 sk 是 user_id，fn
//func NewShardingSelector[T any, SK ShardingKey](sk string,
//	fn ShardingFunc[SK]) *ShardingSelector[T]{
//	return &ShardingSelector[T]{}
//}

//type Int32 int32
//var a ShardingFunc[Int32]



// k 是 sharding key
// fn 就是分库分表的算法
func (s *ShardingSelector[T]) Build() ([]*Query, error) {
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}

	dsts, err := s.findDsts()
	if err != nil {
		return nil, err
	}
	res := make([]*Query, 0, len(dsts))
	for _, dst := range dsts {
		q, err := s.build(dst.DB, dst.Table)
		if err != nil {
			return nil, err
		}
		s.sb = strings.Builder{}
		res = append(res, q)
	}
	return res, nil
}

func (s *ShardingSelector[T]) build(db, tbl string) (*Query, error) {
	s.sb.WriteString("SELECT ")

	if err := s.buildColumns(); err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")

	s.sb.WriteString(fmt.Sprintf("%s.%s", db, tbl))

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL: s.sb.String(),
		Args: s.args,
	}, nil
}

// []Dst: 所有候选的目标节点
// error: 是否出错
func (s *ShardingSelector[T]) findDsts() ([]Dst, error){
	// 在这里，深入（递归）到 WHERE 部分，也就是
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		// 在这里，空切片就意味着，不需要发请求到任何节点
		// WHERE user_id = 123 AND user_id = 124
		return s.findDstByPredicate(p)
	}
	// 这边是要广播
}

// WHERE id = 11 AND user_id = 123
// WHERE user_id = 123 AND id = 11
// WHERE user_id = 123 AND user_id IN (123, 124)
// WHERE user_id = 123 AND user_id = 124
func (s *ShardingSelector[T]) findDstByPredicate(p Predicate) ([]Dst, error) {
	var res []Dst
	switch p.op {
	case opAnd:
		// 空切片意味着广播
		// case1: right 有一个
		// case2: right 是广播
		right, err := s.findDstByPredicate(p.right.(Predicate))
		if err != nil {
			return nil, err
		}
		if len(right) == 0 {
			// 说明广播
			// case2 进来这里
			return s.findDstByPredicate(p.left.(Predicate))
		}
		// case1: left 是广播
		left, err := s.findDstByPredicate(p.left.(Predicate))
		if err != nil {
			return nil, err
		}
		if len(left) ==0 {
			// case1 进来这里
			return right, nil
		}
		// 求交集
		// case 3 进来这里
		return s.merge(left, right), nil
	case opOr:

	//case opLT:
	//case opGT:
	case opEQ:
		left, ok := p.left.(Column)
		if ok {
			_, isSK := s.model.Sks[left.name]
			// WHERE id = 123
			right, ok := p.right.(value)
			if !ok {
				return nil, errors.New("太复杂的查询，暂时不支持")
			}
			if isSK && ok {
				db, tbl := s.model.Sf(right.val)
				res = append(res, Dst{DB: db, Table: tbl})
			}
		}
	default:
		return nil, fmt.Errorf("orm: 不知道怎么处理的操作符")
	}
	return res, nil
}

func (s *ShardingSelector[T]) merge(left, right []Dst) []Dst {
	res := make([]Dst, 0, len(left) + len(right))
	for _, l := range left {
		res = append(res, l)
	}
	for _, r := range right {
		exist := false
		for _, l := range left {
			if r.DB == l.DB && r.Table == l.Table {
				exist = true
			}
		}
		if !exist {
			res = append(res, r)
		}
	}
	return res
}

type Dst struct {
	DB string
	Table string
}

func (s *ShardingSelector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}

	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c)
			if err != nil {
				return err
			}
		case Aggregate:
			// 聚合函数名
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.alias)
				s.sb.WriteByte('`')
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArg(c.args...)
		}
	}

	return nil
}

func (s *ShardingSelector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type){
	case nil:
	case Predicate:
		// 在这里处理 p
		// p.left 构建好
		// p.op 构建好
		// p.right 构建好
		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		if exp.op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(exp.op.String())
			s.sb.WriteByte(' ')
		}
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		// 这种写法很隐晦
		exp.alias = ""
		return s.buildColumn(exp)
	case value:
		s.sb.WriteByte('?')
		s.addArg(exp.val)
	case RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(exp.raw)
		s.addArg(exp.args...)
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}