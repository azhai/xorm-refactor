package crud_test

import (
	"testing"

	"gitee.com/azhai/xorm-refactor/tests/contrib"
	_ "gitee.com/azhai/xorm-refactor/tests/models"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"
	"gitee.com/azhai/xorm-refactor/utils"
	"github.com/stretchr/testify/assert"
)

var allUserData = []map[string]interface{}{
	{
		"username":     "admin",
		"realname":     "管理员",
		"introduction": "不受限的超管账号。",
		"avatar":       "/avatars/avatar-admin.jpg",
	},
	{
		"username":     "demo",
		"realname":     "演示用户",
		"introduction": "演示和测试账号。",
		"avatar":       "/avatars/avatar-demo.jpg",
	},
}

var allGroupData = []map[string]string{
	{"title": "总经办", "remark": "董事会与总经理办公室。"},
	{"title": "技术部", "remark": ""},
	{"title": "行政部", "remark": "行政与人事"},
	{"title": "财务部", "remark": ""},
	{"title": "销售部", "remark": "市场与销售"},
	{"title": "生产部", "remark": "工程与生产"},
}

var userRoleGroups = map[string]map[string][]string{
	"admin": {
		"roles":  {"superuser"},
		"groups": {"总经办", "技术部"},
	},
	"demo": {
		"roles":  {"member"},
		"groups": {"技术部"},
	},
}

// 插入部门表
func insertGroups(t *testing.T, data []map[string]string) map[string]db.Group {
	var groups []db.Group
	for _, row := range data {
		groups = append(groups, db.Group{
			Gid:    utils.NewSerialNo('G'),
			Title:  row["title"],
			Remark: utils.NewNullString(row["remark"]),
		})
	}
	if len(groups) > 0 {
		table := groups[0].TableName()
		// 清空
		err := contrib.TruncTable(table)
		assert.NoError(t, err)
		// 写入
		_, err = db.Table(table).InsertMulti(groups)
		assert.NoError(t, err)
	}
	result := make(map[string]db.Group)
	for _, g := range groups {
		result[g.Title] = g
	}
	return result
}

// 插入用户和角色
func TestMulti01InsertUserRoleGroups(t *testing.T) {
	// 清空
	u, ur := new(db.User), new(db.UserRole)
	err := contrib.TruncTable(u.TableName())
	assert.NoError(t, err)
	err = contrib.TruncTable(ur.TableName())
	assert.NoError(t, err)
	groups := insertGroups(t, allGroupData)

	var userRoles []db.UserRole
	for _, row := range allUserData {
		username := row["username"].(string)
		row["password"] = utils.CreatePassword(username)
		uid := utils.NewSerialNo('U')
		row["uid"] = uid

		uroles := userRoleGroups[username]["roles"]
		for _, rname := range uroles {
			userRoles = append(userRoles, db.UserRole{UserUid: uid, RoleName: rname})
		}

		ugroups := userRoleGroups[username]["groups"]
		for i, gname := range ugroups {
			if grow, ok := groups[gname]; ok {
				if i == 0 {
					row["prin_gid"] = grow.Gid
				} else {
					row["vice_gid"] = grow.Gid
				}
			}
		}

		err = u.Save(row)
		assert.NoError(t, err)
	}

	_, err = db.Table(ur.TableName()).InsertMulti(userRoles)
	assert.NoError(t, err)
}
