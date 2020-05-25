package base

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"xorm.io/xorm"
)

/**
 * 创建或清空表
 */
type IResetTable interface {
	ResetTable(curr string, trunc bool) error
	ITableName
}

/**
 * 写入多行到分布式表中
 */
func ClusterInsertMulti(xi xorm.Interface, objs []IResetTable, reset, trunc bool) (total int64, err error) {
	var count int64
	last, max := 0, len(objs)-1
	for i, m := range objs {
		table := m.TableName()
		if i < max && objs[i+1].TableName() == table {
			continue
		}
		if reset {
			if err = m.ResetTable(table, trunc); err != nil {
				return
			}
		}
		count, err = xi.Table(table).InsertMulti(objs[last : i+1])
		last, total = i+1, total+count
		if err != nil {
			return
		}
	}
	return
}

var clusters = make(map[string]*ClusterMixin)

func GetClusterMixinFor(kind, prefix string, engine *xorm.Engine) *ClusterMixin {
	key := kind + ":" + prefix
	if c, ok := clusters[key]; ok {
		return c
	}
	c := NewClusterMixin(kind, time.Now())
	c.TableNamePrefix = prefix
	c.Suffixes = FindTables(engine, prefix, false)
	c.Suffixes.Sort()
	clusters[key] = c
	return c
}

// 按月分表
type ClusterMixin struct {
	Date            time.Time
	Suffixes        sort.StringSlice
	TableNamePrefix string
	Kind, Format    string
}

func NewClusterMixin(kind string, t time.Time) *ClusterMixin {
	m := &ClusterMixin{Kind: kind}
	return m.SetTime(t, true)
}

func NewClusterQuarterly(t time.Time) *ClusterMixin {
	return NewClusterMixin("Quarterly", t)
}

func NewClusterMonthly(t time.Time) *ClusterMixin {
	return NewClusterMixin("Monthly", t)
}

func NewClusterWeekly(t time.Time) *ClusterMixin {
	return NewClusterMixin("Weekly", t)
}

func NewClusterDaily(t time.Time) *ClusterMixin {
	return NewClusterMixin("Daily", t)
}

func NewClusterHourly(t time.Time) *ClusterMixin {
	return NewClusterMixin("Hourly", t)
}

func (m ClusterMixin) GetSuffix() string {
	if !strings.Contains(m.Format, "%") {
		return m.Date.Format(m.Format)
	}
	s := m.Date.Format("2006010215")
	year, week := m.Date.ISOWeek()
	yearday := m.Date.YearDay()
	weekday := int(m.Date.Weekday())
	quarter := int(m.Date.Month()+2) / 3
	repl := strings.NewReplacer(
		"%Y", s[:4], "%m", s[4:6], "%d", s[6:8], "%H", s[8:],
		"%R", strconv.Itoa(year), "%W", strconv.Itoa(week),
		"%D", strconv.Itoa(yearday), "%w", strconv.Itoa(weekday),
		"%Q", strconv.Itoa(quarter),
	)
	return repl.Replace(m.Format)
}

func (m *ClusterMixin) SetTime(t time.Time, toFirst bool) *ClusterMixin {
	if !toFirst || t.IsZero() {
		m.Date = t
		return m
	}
	loc := time.Local
	year, month, day := t.Date()
	switch m.Kind {
	case "Quarterly":
		month = (month + 2) / 3
		m.Date = time.Date(year, month, 1, 0, 0, 0, 0, loc)
		m.Format = "%Y%Q"
	case "Monthly":
		m.Date = time.Date(year, month, 1, 0, 0, 0, 0, loc)
		m.Format = "200601"
	case "Weekly":
		day = day - int(t.Weekday())
		m.Date = time.Date(year, month, day, 0, 0, 0, 0, loc)
		m.Format = "%R0%W"
	case "Daily":
		m.Date = time.Date(year, month, day, 0, 0, 0, 0, loc)
		m.Format = "20060102"
	case "Hourly":
		m.Date = time.Date(year, month, day, t.Hour(), 0, 0, 0, loc)
		m.Format = "2006010215"
	}
	return m
}

func (m *ClusterMixin) Move(n int) *ClusterMixin {
	switch m.Kind {
	case "Quarterly":
		m.Date = m.Date.AddDate(0, n*3, 0)
	case "Monthly":
		m.Date = m.Date.AddDate(0, n, 0)
	case "Weekly":
		m.Date = m.Date.AddDate(0, 0, n*7)
	case "Daily":
		m.Date = m.Date.AddDate(0, 0, n)
	case "Hourly":
		m.Date = m.Date.Add(time.Hour * time.Duration(n))
	}
	return m
}

func (m *ClusterMixin) Prev() bool {
	suffix := m.GetSuffix()
	idx := m.Suffixes.Search(suffix)
	if idx < 1 || suffix != m.Suffixes[idx] {
		return false
	}
	target := m.Suffixes[idx-1]
	for m.GetSuffix() > target {
		m.Move(-1)
	}
	return m.GetSuffix() == target
}

func (m *ClusterMixin) Next() bool {
	suffix := m.GetSuffix()
	idx := m.Suffixes.Search(suffix)
	max := m.Suffixes.Len() - 1
	if idx < 0 || idx >= max || suffix != m.Suffixes[idx] {
		return false
	}
	target := m.Suffixes[idx+1]
	for m.GetSuffix() < target {
		m.Move(1)
	}
	return m.GetSuffix() == target
}

// 分布式查询
type ClusterQuery struct {
	engine  *xorm.Engine
	filters []FilterFunc
	*ClusterMixin
	*xorm.Session
}

func NewClusterQuery(engine *xorm.Engine, cluster *ClusterMixin) *ClusterQuery {
	table := cluster.TableNamePrefix + cluster.GetSuffix()
	return &ClusterQuery{
		engine:       engine,
		filters:      nil,
		ClusterMixin: cluster,
		Session:      engine.Table(table),
	}
}

func (q ClusterQuery) Quote(value string) string {
	return q.engine.Quote(value)
}

func (q *ClusterQuery) ClearFilters() *ClusterQuery {
	q.filters = make([]FilterFunc, 0)
	return q
}

func (q *ClusterQuery) AddFilter(filter FilterFunc) *ClusterQuery {
	q.filters = append(q.filters, filter)
	return q
}

func (q *ClusterQuery) Limit(limit int, start ...int) *ClusterQuery {
	q.AddFilter(func(query *xorm.Session) *xorm.Session {
		return query.Limit(limit, start...)
	})
	return q
}

func (q *ClusterQuery) OrderBy(order string) *ClusterQuery {
	q.AddFilter(func(query *xorm.Session) *xorm.Session {
		return query.OrderBy(order)
	})
	return q
}

// 获取当前表名
func (q *ClusterQuery) GetTable() string {
	return q.TableNamePrefix + q.GetSuffix()
}

// 重新构建当前查询，因为每次 COUNT 和 FIND 等操作会释放查询（只有主表名还保留着）
func (q *ClusterQuery) GetQuery() *xorm.Session {
	query := q.Session.Clone()
	query = query.Table(q.GetTable())
	for _, filter := range q.filters {
		query = filter(query)
	}
	return query
}

// 计数，只查询当前表
func (q *ClusterQuery) Count(bean ...interface{}) (int64, error) {
	return q.GetQuery().Count(bean...)
}

// 计数和获取结果集，只查询当前表
func (q *ClusterQuery) FindAndCount(rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
	return q.GetQuery().FindAndCount(rowsSlicePtr, condiBean...)
}

func (q *ClusterQuery) clusterCounts(bean ...interface{}) (map[string]int64, error) {
	counts := make(map[string]int64)
	if q.Suffixes.Len() == 0 {
		return counts, nil
	}
	q.SetTime(time.Now(), true)
	for ok := true; ok; ok = q.Prev() {
		count, err := q.GetQuery().Count(bean...)
		if err != nil {
			return counts, err
		}
		counts[q.GetTable()] = count
	}
	return counts, nil
}

// 分布式计数，查询全部表
func (q *ClusterQuery) ClusterCount(bean ...interface{}) (int64, error) {
	counts, err := q.clusterCounts(bean...)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, count := range counts {
		total += count
	}
	return total, nil
}

// 分布式计数和翻页，查询全部表，但只获取部分结果集
func (q *ClusterQuery) ClusterPaginate(pageno, pagesize int,
	rowsSlicePtr interface{}, condiBean ...interface{}) (int64, error) {
	if q.Suffixes.Len() == 0 {
		return 0, nil
	}
	q.SetTime(time.Now(), true)
	var total int64
	limit, offset := CalcPage(pageno, pagesize, 0)
	for ok := true; ok; ok = q.Prev() {
		count, err := q.GetQuery().Count()
		if err != nil {
			return total, err
		}
		total += count
		if limit <= 0 || count == 0 {
			continue
		}
		if offset >= int(count) {
			offset -= int(count)
		} else {
			query := q.GetQuery().Limit(limit, offset)
			err = query.Find(rowsSlicePtr, condiBean...)
			limit, offset = limit+offset-int(count), 0
		}
	}
	return total, nil
}
