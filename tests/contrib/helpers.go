package contrib

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/azhai/xorm-refactor/builtin/access"
	"gitee.com/azhai/xorm-refactor/builtin/base"
	"gitee.com/azhai/xorm-refactor/tests/models/default"
)

// 清空表
func TruncTable(tableName string) error {
	engine := db.Engine()
	sql := base.Qprintf(engine, "TRUNCATE TABLE %s", tableName)
	_, err := engine.Exec(sql)
	return err
}

func CountRows(tableName string, excludeDeleted bool) int {
	query := db.Table(tableName)
	if excludeDeleted {
		column := db.Quote("deleted_at")
		query.Where(fmt.Sprintf("%s IS NULL", column))
	}
	total, err := query.Count()
	if err != nil {
		return -1
	}
	return int(total)
}

func GetUserInfo(user *db.User) map[string]interface{} {
	info := map[string]interface{}{
		"username": user.Username,
	}
	if user.Id > 0 {
		info["id"] = strconv.Itoa(user.Id)
	}
	if user.Realname != "" {
		info["realname"] = user.Realname
	}
	if user.Mobile != "" {
		info["mobile"] = user.Mobile
	}
	if user.Avatar != "" {
		info["avatar"] = user.Avatar
	}
	if user.Introduction != "" {
		info["introduction"] = user.Introduction
	}
	return info
}

func GetUserRoles(user *db.User) (roles []string, err error) {
	if user.Uid == "" {
		return
	}
	query := db.Table(db.UserRole{}).Cols("role_name")
	err = query.Where("user_uid = ?", user.Uid).Find(&roles)
	return
}

func NewMenu(path, title, icon string) *db.Menu {
	return &db.Menu{
		Path: path, Title: title, Icon: icon,
	}
}

// 添加子菜单
func AddMenuToParent(menu, parent *db.Menu) (err error) {
	var parentNode *base.NestedMixin
	if parent != nil {
		parentNode = parent.NestedMixin
	}
	if menu.NestedMixin == nil {
		menu.NestedMixin = new(base.NestedMixin)
	}
	query, table := db.Table(), menu.TableName()
	err = menu.NestedMixin.AddToParent(parentNode, query, table)
	if err == nil {
		_, err = query.InsertOne(menu)
	}
	return
}

// 添加权限
func AddAccess(role, res string, perm uint16, args ...string) (acc *db.Access, err error) {
	acc = &db.Access{
		RoleName: role, PermCode: int(perm),
		ResourceType: res, GrantedAt: time.Now(),
	}
	_, names := access.ParsePermNames(uint16(acc.PermCode))
	acc.Actions = strings.Join(names, ",")
	if len(args) > 0 {
		resArgs := strings.Join(args, ",")
		acc.ResourceArgs = resArgs
	}
	_, err = db.Table().InsertOne(acc)
	return
}
