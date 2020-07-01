package builtin

import (
	"bytes"
	"fmt"

	"gitee.com/azhai/xorm-refactor/builtin/sqljoin"
	"gitee.com/azhai/xorm-refactor/utils"
	"xorm.io/xorm"
)

// 联表查询
func JoinQuery(engine *xorm.Engine, query *xorm.Session,
	table, fkey string, foreign ForeignTable) (*xorm.Session, []string) {
	frgTable, frgAlias := foreign.TableName(), foreign.AliasName()
	cond := Qprintf(engine, "%s.%s = %s.%s", table, fkey, frgAlias, foreign.Index)
	if query == nil {
		query = engine.Table(table)
	}
	var cols []string
	cols = utils.GetColumns(foreign.Table, frgAlias, cols)
	query = query.Join(foreign.Join.Subject(), frgTable, cond)
	return query, cols
}

// 关联表
type ForeignTable struct {
	Join  sqljoin.SqlJoin
	Table ITableName
	Alias string
	Index string
}

// 表名或别名，通常用于字段之前
func (f ForeignTable) AliasName() string {
	if f.Alias != "" {
		return f.Alias
	}
	return f.Table.TableName()
}

// 表名和别名，通常用于 FROM 或 Join 之后
func (f ForeignTable) TableName() string {
	table := f.Table.TableName()
	if f.Alias != "" {
		return fmt.Sprintf("%s AS %s", table, f.Alias)
	}
	return table
}

// Left Join 联表查询
type LeftJoinQuery struct {
	engine      *xorm.Engine
	filters     []FilterFunc
	nativeTable string
	Native      ITableName
	ForeignKeys []string
	Foreigns    map[string]ForeignTable
	*xorm.Session
}

// native 为最左侧的主表，查询其所有字段
func NewLeftJoinQuery(engine *xorm.Engine, native ITableName) *LeftJoinQuery {
	nativeTable := native.TableName()
	return &LeftJoinQuery{
		engine:      engine,
		filters:     nil,
		nativeTable: nativeTable,
		Native:      native,
		Foreigns:    make(map[string]ForeignTable),
		Session:     engine.Table(nativeTable),
	}
}

func (q LeftJoinQuery) Quote(value string) string {
	return q.engine.Quote(value)
}

func (q *LeftJoinQuery) ClearFilters() *LeftJoinQuery {
	q.filters = make([]FilterFunc, 0)
	return q
}

func (q *LeftJoinQuery) AddFilter(filter FilterFunc) *LeftJoinQuery {
	q.filters = append(q.filters, filter)
	return q
}

// foreign 为副表，只查询其部分字段，读取字段的 json tag 作为字段名
func (q *LeftJoinQuery) LeftJoin(foreign ITableName, fkey string) *LeftJoinQuery {
	q.AddLeftJoin(foreign, "", fkey, "")
	return q
}

// 添加次序要和 struct 定义一致
func (q *LeftJoinQuery) AddLeftJoin(foreign ITableName, pkey, fkey, alias string) *LeftJoinQuery {
	if pkey == "" {
		col := GetPrimarykey(q.engine, foreign)
		if col != nil {
			pkey = col.Name
		}
	}
	if _, ok := q.Foreigns[fkey]; !ok {
		q.ForeignKeys = append(q.ForeignKeys, fkey)
	}
	q.Foreigns[fkey] = ForeignTable{
		Join:  sqljoin.LeftJoin,
		Table: foreign,
		Alias: alias,
		Index: pkey,
	}
	return q
}

func (q *LeftJoinQuery) Limit(limit int, start ...int) *LeftJoinQuery {
	q.AddFilter(func(query *xorm.Session) *xorm.Session {
		return query.Limit(limit, start...)
	})
	return q
}

func (q *LeftJoinQuery) OrderBy(order string) *LeftJoinQuery {
	q.AddFilter(func(query *xorm.Session) *xorm.Session {
		return query.OrderBy(order)
	})
	return q
}

// 重新构建当前查询，因为每次 COUNT 和 FIND 等操作会释放查询（只有主表名还保留着）
func (q *LeftJoinQuery) GetQuery() *xorm.Session {
	buf := new(bytes.Buffer)
	buf.WriteString(Qprintf(q.engine, "%s.*", q.Native.TableName()))
	query := q.Session
	for _, filter := range q.filters {
		query = filter(query)
	}
	var cols []string
	for _, fkey := range q.ForeignKeys {
		foreign := q.Foreigns[fkey]
		query, cols = JoinQuery(q.engine, query, q.nativeTable, fkey, foreign)
		buf.WriteString(", ")
		buf.WriteString(utils.QuoteColumns(cols, ", ", q.engine.Quote))
	}
	return query.Select(buf.String())
}

// 计数，由于左联接数量只跟主表有关，这里不去 Join
func (q *LeftJoinQuery) Count(bean ...interface{}) (int64, error) {
	query := q.Session
	for _, filter := range q.filters {
		query = filter(query)
	}
	return query.Count(bean...)
}

// 计数和获取结果集
func (q *LeftJoinQuery) FindAndCount(
	rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
	total, err := q.Count()
	if err != nil || total == 0 {
		return total, err
	}
	err = q.GetQuery().Find(rowsSlicePtr, condiBean...)
	return total, err
}

// 计数和翻页，只获取部分结果集
func (q *LeftJoinQuery) FindPaginate(pageno, pagesize int,
	rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
	total, err := q.Count()
	limit, offset := CalcPage(pageno, pagesize, int(total))
	query := q.GetQuery()
	if limit >= 0 {
		query = query.Limit(limit, offset)
	}
	err = query.Find(rowsSlicePtr, condiBean...)
	return total, err
}
