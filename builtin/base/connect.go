package base

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// 过滤查询
type FilterFunc = func(query *xorm.Session) *xorm.Session

// 修改操作，用于事务
type ModifyFunc = func(tx *xorm.Session) (int64, error)

/**
 * 数据表名
 */
type ITableName interface {
	TableName() string
}

/**
 * 数据表注释
 */
type ITableComment interface {
	TableComment() string
}

// 对参数先进行转义Quote
func Qprintf(engine *xorm.Engine, format string, args ...interface{}) string {
	if engine != nil {
		for i, arg := range args {
			args[i] = engine.Quote(arg.(string))
		}
	}
	return fmt.Sprintf(format, args...)
}

// 盲转义，认定字段名以小写字母开头
func BlindlyQuote(engine *xorm.Engine, sep string, words ...string) string {
	repl := engine.Quote("$1")
	origin := strings.Join(words, sep)
	re := regexp.MustCompile("([a-z][a-zA-Z0-9_]+)")
	result := re.ReplaceAllString(origin, repl)
	if pad := (len(repl) - len("$1")) / 2; pad > 0 {
		left, right := repl[:pad], repl[len(repl)-pad:]
		oldnew := []string{
			left + left, left, right + right, right,
			"'" + left, "'", left + "'", "'",
		}
		result = strings.NewReplacer(oldnew...).Replace(result)
	}
	return result
}

// 找出符合前缀的表名
func FindTables(engine *xorm.Engine, prefix string, fullName bool) []string {
	var result []string
	db, ctx := engine.DB(), context.Background()
	tables, err := engine.Dialect().GetTables(db, ctx)
	if err != nil {
		return result
	}
	prelen := len(prefix)
	for _, t := range tables {
		if prelen > 0 && !strings.HasPrefix(t.Name, prefix) {
			continue
		}
		if fullName {
			result = append(result, t.Name)
		} else {
			result = append(result, t.Name[prelen:])
		}
	}
	return result
}

// 复制表结构，只用于MySQL
func CreateTableLike(engine *xorm.Engine, curr, orig string) (bool, error) {
	if engine.DriverName() != "mysql" {
		err := fmt.Errorf("Only support mysql/mariadb database !")
		return false, err
	}
	exists, err := engine.IsTableExist(curr)
	if err != nil || exists {
		return false, err
	}
	sql := "CREATE TABLE IF NOT EXISTS %s LIKE %s"
	_, err = engine.Exec(Qprintf(engine, sql, curr, orig))
	return err == nil, err
}

// 获取Model的字段列表
func GetPrimarykey(engine *xorm.Engine, m interface{}) *schemas.Column {
	table, err := engine.TableInfo(m)
	if err != nil {
		return nil
	}
	if cols := table.PKColumns(); len(cols) > 0 {
		return cols[0]
	}
	return nil
}

// 调整从后往前翻页
func NegativeOffset(offset, pagesize, total int) int {
	if remain := total % pagesize; remain > 0 {
		offset += pagesize - remain
	}
	return offset + total
}

// 计算翻页
func CalcPage(pageno, pagesize, total int) (int, int) {
	if pagesize < 0 {
		return -1, 0
	} else if pagesize == 0 {
		return 0, 0
	}
	var offset int
	if pageno > 0 {
		offset = (pageno - 1) * pagesize
	} else if pageno < 0 && total > 0 {
		offset = NegativeOffset(pageno*pagesize, pagesize, total)
	}
	return pagesize, offset
}

// 使用翻页
func Paginate(query *xorm.Session, pageno, pagesize int) *xorm.Session {
	var limit, offset int
	if pagesize > 0 && pageno < 0 {
		total, _ := query.Count()
		limit, offset = CalcPage(pageno, pagesize, int(total))
	} else {
		limit, offset = CalcPage(pageno, pagesize, 0)
	}
	if limit >= 0 {
		query = query.Limit(limit, offset)
	}
	return query
}

/**
 * 时间相关的三个典型字段
 */
type TimeMixin struct {
	CreatedAt time.Time `json:"created_at" xorm:"created comment('创建时间') TIMESTAMP"`       // 创建时间
	UpdatedAt time.Time `json:"updated_at" xorm:"updated comment('更新时间') TIMESTAMP"`       // 更新时间
	DeletedAt time.Time `json:"deleted_at" xorm:"deleted comment('删除时间') index TIMESTAMP"` // 删除时间
}
