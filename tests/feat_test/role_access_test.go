package feat_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/base"
	"gitee.com/azhai/xorm-refactor/enums"

	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"
	"github.com/stretchr/testify/assert"
)

func TestAccess01Insert(t *testing.T) {
	m := new(db.Access)
	err := contrib.TruncTable(m.TableName())
	assert.NoError(t, err)
	// 超管可以访问所有菜单
	allActs := enums.Permit(2<<16 - 1)
	m, err = contrib.AddAccess("superuser", "menu", allActs, "*")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 1)
	// 普通用户
	userActs := enums.View | enums.Add | enums.Edit
	m, err = contrib.AddAccess("member", "menu", userActs, "/dashboard")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 2)
	m, err = contrib.AddAccess("member", "menu", enums.View, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 3)
	// 未登录用户
	m, err = contrib.AddAccess("", "menu", enums.View, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 4)
}

func TestAccess02Anonymous(t *testing.T) {
	var err error
	anonymous := &contrib.UserAuth{}
	err = base.Authorize(anonymous, enums.View, "style.css")
	assert.NoError(t, err)
	err = base.Authorize(anonymous, enums.Edit, "/images/abc.jpg")
	assert.NoError(t, err)
	err = base.Authorize(anonymous, enums.Edit, "/error/404")
	assert.NoError(t, err)
	err = base.Authorize(anonymous, enums.Edit, "/dashboard")
	assert.Error(t, err) // 无权限！
}

func TestAccess03Demo(t *testing.T) {
	var err error
	demo := &contrib.UserAuth{User: new(db.User)}
	demo.User.Load("username = ?", "demo")
	err = base.Authorize(demo, enums.Draft, "/images/abc.jpg")
	assert.NoError(t, err)
	err = base.Authorize(demo, enums.Edit, "/dashboard")
	assert.NoError(t, err)
	err = base.Authorize(demo, enums.View, "/notExists")
	assert.Error(t, err) // 无权限！
}

func TestAccess04Admin(t *testing.T) {
	var err error
	admin := &contrib.UserAuth{User: new(db.User)}
	admin.User.Load("username = ?", "admin")
	err = base.Authorize(admin, enums.Edit, "/images/abc.jpg")
	assert.NoError(t, err)
	err = base.Authorize(admin, enums.Delete, "/dashboard")
	assert.NoError(t, err)
	err = base.Authorize(admin, enums.Add, "/notExists")
	assert.NoError(t, err)
	err = base.Authorize(admin, enums.Noop, "")
	assert.NoError(t, err)
}
