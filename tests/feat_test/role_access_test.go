package feat_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/builtin/access"
	"gitee.com/azhai/xorm-refactor/builtin/auth"
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
	m, err = contrib.AddAccess("superuser", "menu", access.ALL, "*")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 1)
	// 普通用户
	ops := access.VIEW | access.GET | access.POST
	m, err = contrib.AddAccess("member", "menu", ops, "/dashboard")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 2)
	m, err = contrib.AddAccess("member", "menu", access.VIEW, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 3)
	// 未登录用户
	m, err = contrib.AddAccess("", "menu", access.VIEW, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 4)
}

func TestAccess02Anonymous(t *testing.T) {
	var err error
	anonymous := &contrib.UserAuth{}
	err = auth.Authorize(anonymous, access.VIEW, "style.css")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, access.POST, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, access.POST, "/error/404")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, access.POST, "/dashboard")
	assert.Error(t, err) // 无权限！
}

func TestAccess03Demo(t *testing.T) {
	var err error
	demo := &contrib.UserAuth{User: new(db.User)}
	demo.User.Load("username = ?", "demo")
	err = auth.Authorize(demo, access.DISABLE, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(demo, access.POST, "/dashboard")
	assert.NoError(t, err)
	err = auth.Authorize(demo, access.VIEW, "/notExists")
	assert.Error(t, err) // 无权限！
}

func TestAccess04Admin(t *testing.T) {
	var err error
	admin := &contrib.UserAuth{User: new(db.User)}
	admin.User.Load("username = ?", "admin")
	err = auth.Authorize(admin, access.POST, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(admin, access.REMOVE, "/dashboard")
	assert.NoError(t, err)
	err = auth.Authorize(admin, access.GET, "/notExists")
	assert.NoError(t, err)
	err = auth.Authorize(admin, access.NONE, "")
	assert.NoError(t, err)
}
