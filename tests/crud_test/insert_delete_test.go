package crud_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	"gitee.com/azhai/xorm-refactor/tests/models/default"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

var allRoleData = []map[string]interface{}{
	{"name": "superuser", "remark": "超级用户，无上权限的超级管理员。"},
	{"name": "member", "remark": "普通用户，除权限外的其他页面。"},
	{"name": "visitor", "remark": "基本用户，只能看到面板页。"},
}

// 插入三个角色
func TestSingle01InsertRoles(t *testing.T) {
	m := new(db.Role)
	err := contrib.TruncTable(m.TableName())
	assert.NoError(t, err)
	for _, row := range allRoleData {
		err = m.Save(row)
		assert.NoError(t, err)
		assert.Equal(t, m.Id, 0)
	}
}

// 软删除第二个角色
func TestSingle02SoftDeleteRole(t *testing.T) {
	m := &db.Role{Id: 2}
	table := m.TableName()
	assert.Equal(t, contrib.CountRows(table, true), 3)
	err := db.ExecTx(func(tx *xorm.Session) (int64, error) {
		return tx.Delete(m)
	})
	assert.NoError(t, err)
	assert.Equal(t, contrib.CountRows(table, true), 2)
	assert.Equal(t, contrib.CountRows(table, false), 3)
}
