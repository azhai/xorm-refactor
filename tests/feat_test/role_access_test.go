package feat_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/builtin/auth"
	"gitee.com/azhai/xorm-refactor/builtin/userpermit"
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
	allActs := userpermit.UserPermit(2<<16 - 1)
	m, err = contrib.AddAccess("superuser", "menu", allActs, "*")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 1)
	// 普通用户
	userActs := userpermit.View | userpermit.Add | userpermit.Edit
	m, err = contrib.AddAccess("member", "menu", userActs, "/dashboard")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 2)
	m, err = contrib.AddAccess("member", "menu", userpermit.View, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 3)
	// 未登录用户
	m, err = contrib.AddAccess("", "menu", userpermit.View, "/error/404")
	assert.NoError(t, err)
	assert.Equal(t, m.Id, 4)
}

func TestAccess02Anonymous(t *testing.T) {
	var err error
	anonymous := &contrib.UserAuth{}
	err = auth.Authorize(anonymous, userpermit.View, "style.css")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, userpermit.Edit, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, userpermit.Edit, "/error/404")
	assert.NoError(t, err)
	err = auth.Authorize(anonymous, userpermit.Edit, "/dashboard")
	assert.Error(t, err) // 无权限！
}

func TestAccess03Demo(t *testing.T) {
	var err error
	demo := &contrib.UserAuth{User: new(db.User)}
	demo.User.Load("username = ?", "demo")
	err = auth.Authorize(demo, userpermit.Draft, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(demo, userpermit.Edit, "/dashboard")
	assert.NoError(t, err)
	err = auth.Authorize(demo, userpermit.View, "/notExists")
	assert.Error(t, err) // 无权限！
}

func TestAccess04Admin(t *testing.T) {
	var err error
	admin := &contrib.UserAuth{User: new(db.User)}
	admin.User.Load("username = ?", "admin")
	err = auth.Authorize(admin, userpermit.Edit, "/images/abc.jpg")
	assert.NoError(t, err)
	err = auth.Authorize(admin, userpermit.Delete, "/dashboard")
	assert.NoError(t, err)
	err = auth.Authorize(admin, userpermit.Add, "/notExists")
	assert.NoError(t, err)
	err = auth.Authorize(admin, userpermit.Noop, "")
	assert.NoError(t, err)
}
