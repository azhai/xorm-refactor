package feat_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/cmd"
	"github.com/k0kubun/pp"

	"gitee.com/azhai/xorm-refactor/builtin/base"
	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"

	"github.com/stretchr/testify/assert"
)

var allMenuData = []map[string]string{
	{"path": "/dashboard", "title": "面板", "icon": "dashboard"},
	{"path": "/permission", "title": "权限", "icon": "lock"},
	{"path": "role", "title": "角色权限"},
	{"path": "/table", "title": "Table", "icon": "table"},
	{"path": "complex-table", "title": "复杂Table"},
	{"path": "inline-edit-table", "title": "内联编辑"},
	{"path": "/excel", "title": "Excel", "icon": "excel"},
	{"path": "export-selected-excel", "title": "选择导出"},
	{"path": "upload-excel", "title": "上传Excel"},
	{"path": "/theme/index", "title": "主题", "icon": "theme"},
	{"path": "/error/404", "title": "404错误", "icon": "404"},
	{"path": "https://cn.vuejs.org/", "title": "外部链接", "icon": "link"},
}

type Menu struct {
	db.Menu  `json:",inline" xorm:"extends"`
	Children []Menu `json:"children" xorm:"-"`
}

// 添加菜单
func TestNested01InsertMenus(t *testing.T) {
	parent := new(db.Menu)
	err := contrib.TruncTable(parent.TableName())
	assert.NoError(t, err)
	icon, ok := "", false
	for _, row := range allMenuData {
		if icon, ok = row["icon"]; ok && icon != "" { // 第一级顶级菜单，都有icon，路径以斜线开头
			parent = contrib.NewMenu(row["path"], row["title"], icon)
			err := contrib.AddMenuToParent(parent, nil)
			assert.NoError(t, err)
		} else { // 第二级子菜单，没有icon，属于最近的一个有icon的顶级菜单
			menu := contrib.NewMenu(row["path"], row["title"], icon)
			err := contrib.AddMenuToParent(menu, parent)
			assert.NoError(t, err)
		}
	}
}

// 重建左右端点
func TestNested02Rebuild(t *testing.T) {
	table := (db.Menu{}).TableName()
	count, err := db.Table(table).Count()
	assert.NoError(t, err)
	var affects int64
	changes := map[string]interface{}{"lft": 0, "rgt": 0}
	affects, err = db.Table(table).Update(changes)
	assert.NoError(t, err)
	assert.Equal(t, count, affects)
	err = base.RebuildNestedByDepth(db.Table(), table)
	assert.NoError(t, err)
	count, err = db.Table(table).Where("lft = 0 OR rgt = 0").Count()
	assert.NoError(t, err)
	assert.Equal(t, count, int64(0))
}

// 查找祖先菜单
func TestNested03FindAncestors(t *testing.T) {
	menu := new(db.Menu)
	_, err := menu.Load("path = ?", "inline-edit-table")
	assert.NoError(t, err)
	filter := menu.AncestorsFilter(true)
	var menus []*db.Menu
	err = filter(db.Table(menu)).Find(&menus)
	assert.NoError(t, err)
	if assert.Len(t, menus, 1) {
		assert.Equal(t, menus[0].Path, "/table")
	}
}

// 查找子菜单
func TestNested04FindChildren(t *testing.T) {
	menu := new(db.Menu)
	_, err := menu.Load("path = ?", "/excel")
	assert.NoError(t, err)
	filter := menu.ChildrenFilter(-1, true)
	var menus []*db.Menu
	err = filter(db.Table(menu)).Find(&menus)
	assert.NoError(t, err)
	if assert.Len(t, menus, 2) {
		assert.Equal(t, menus[0].Path, "export-selected-excel")
		assert.Equal(t, menus[1].Path, "upload-excel")
	}
}

// 查找所有菜单
func TestNested05AllMenus(t *testing.T) {
	var menus []Menu
	filter := new(base.NestedMixin).ChildrenFilter(-1, false)
	err := db.QueryAll(filter).Find(&menus)
	assert.NoError(t, err)

	rank, rankChildren := 0, make(map[int][]Menu, 0)
	for _, m := range menus {
		if _, ok := rankChildren[m.Depth]; !ok {
			rankChildren[m.Depth] = make([]Menu, 0)
		}
		if m.Depth < rank {
			m.Children = rankChildren[rank]
			rankChildren[rank] = make([]Menu, 0)
		} else {
			m.Children = make([]Menu, 0)
		}
		rank = m.Depth
		rankChildren[rank] = append(rankChildren[rank], m)
	}
	menus = rankChildren[rank] // 组装好的树状菜单

	if cmd.Verbose() {
		pp.Println(menus)
	}
	if assert.Len(t, menus, 7) {
		assert.Equal(t, menus[3].Path, "/excel")
		children := menus[3].Children
		pp.Println(children)
		assert.Equal(t, children[1].Path, "upload-excel")
	}
}
