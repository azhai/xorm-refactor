package crud_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/builtin"
	"gitee.com/azhai/xorm-refactor/builtin/sqljoin"
	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"
	"gitee.com/azhai/xorm-refactor/utils"
	"github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

func getFilter(engine *xorm.Engine, table string) builtin.FilterFunc {
	return func(query *xorm.Session) *xorm.Session {
		cond := builtin.Qprintf(engine, "%s.%s IS NOT NULL", table, "prin_gid")
		sort := builtin.Qprintf(engine, "%s.%s ASC", table, "id")
		return query.Where(cond).OrderBy(sort)
	}
}

func TestJoin01FindUserGroups(t *testing.T) {
	m := &contrib.UserWithGroup{
		PrinGroup: new(contrib.GroupSummary),
	}
	engine, table := db.Engine(), m.TableName()
	filter := getFilter(engine, table)

	query := db.Table(table)
	total, err := filter(query).Count()
	assert.NoError(t, err)
	var objs []*contrib.UserWithGroup
	if err == nil && total > 0 {
		var cols []string
		cols = utils.GetColumns(m.User, m.User.TableName(), cols)
		query = filter(query).Cols(cols...)
		if testing.Verbose() {
			pp.Println(cols)
		}

		foreign := builtin.ForeignTable{
			Join:  sqljoin.LeftJoin,
			Table: *m.PrinGroup,
			Alias: "", Index: "gid",
		}
		foreign.Alias = "P"
		query, cols = builtin.JoinQuery(engine, query, table, "prin_gid", foreign)
		query = query.Cols(cols...)
		if testing.Verbose() {
			pp.Println(cols)
		}
		foreign.Alias = "V"
		query, cols = builtin.JoinQuery(engine, query, table, "vice_gid", foreign)
		query = query.Cols(cols...)
		if testing.Verbose() {
			pp.Println(cols)
		}

		pageno, pagesize := 1, 20
		builtin.Paginate(query, pageno, pagesize).Find(&objs)
		if testing.Verbose() {
			pp.Println(objs)
		}
	}
}

func TestJoin02LeftJoinQuery(t *testing.T) {
	engine, native := db.Engine(), db.User{}
	table := native.TableName()
	filter := getFilter(engine, table)

	group := contrib.GroupSummary{}
	query := builtin.NewLeftJoinQuery(engine, native)
	query.AddLeftJoin(group, "gid", "prin_gid", "P")
	query.AddLeftJoin(group, "gid", "vice_gid", "V")

	var objs []*contrib.UserWithGroup
	pageno, pagesize := 1, 20
	query = query.AddFilter(filter)
	_, err := query.FindPaginate(pageno, pagesize, &objs)
	assert.NoError(t, err)
	if testing.Verbose() {
		pp.Println(objs)
	}
}
