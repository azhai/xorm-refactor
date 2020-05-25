package contrib

import (
	"strings"

	"gitee.com/azhai/xorm-refactor/builtin/auth"
	"gitee.com/azhai/xorm-refactor/builtin/usertype"
	db "gitee.com/azhai/xorm-refactor/tests/models/default"
	"github.com/azhai/gozzo-utils/common"
	"xorm.io/xorm"
)

const (
	ROLE_NAME_SUPER   = "superuser" // 超级管理员
	ROLE_NAME_LIMITED = "limited"   // 受限用户
	URL_PREFIX_IMAGE  = "/images"   // 静态图片网址
)

var (
	allOpenUrls, limitedBlackList, limitedWhiteList  []string
	loadedOpenUrls, loadedBlackList, loadedWhiteList bool
)

type UserAuth struct {
	*db.User
	auth.IUserAuth
}

// 用户分类，无法区分内部用户和普通用户
func (a UserAuth) GetUserType() (utype usertype.UserType, err error) {
	if a.User == nil || a.User.Id == 0 {
		utype = usertype.ANONYMOUS
		return
	}
	if !a.User.DeletedAt.IsZero() {
		utype = usertype.FORBIDDEN
		return
	}
	var roles []string
	roles, err = a.GetUserRoles()
	if common.InStringList(ROLE_NAME_LIMITED, roles) {
		utype = usertype.LIMITED
	} else if common.InStringList(ROLE_NAME_SUPER, roles) {
		utype = usertype.SUPER
	} else {
		utype = usertype.REGULAR
	}
	return
}

// 用户拥有的角色
func (a UserAuth) GetUserRoles() (roles []string, err error) {
	return GetUserRoles(a.User)
}

// 是否静态资源网址
func (a UserAuth) IsStaticResourceUrl(url string) bool {
	return strings.HasPrefix(url, URL_PREFIX_IMAGE) ||
		strings.HasSuffix(url, ".css") || strings.HasSuffix(url, ".js")
}

// 获取可公开访问的网址
func (a UserAuth) GetAnonymousOpenUrls() []string {
	if loadedOpenUrls {
		return allOpenUrls
	}
	query := QueryPermissions().Where("perm_code > 0")
	query.Cols("resource_args").Find(&allOpenUrls)
	loadedOpenUrls = true
	return allOpenUrls
}

// 获取受限用户黑名单中的的网址，与白名单二选一
func (a UserAuth) GetLimitedBlackListUrls() []string {
	if loadedBlackList {
		return limitedBlackList
	}
	query := QueryPermissions(ROLE_NAME_LIMITED).Where("perm_code = 0")
	query.Cols("resource_args").Find(&limitedBlackList)
	loadedBlackList = true
	return limitedBlackList
}

// 获取受限用户白名单中的的网址，不再检查正常用户权限，与黑名单二选一
func (a UserAuth) GetLimitedWhiteListUrls() []string {
	if loadedWhiteList {
		return limitedWhiteList
	}
	query := QueryPermissions(ROLE_NAME_LIMITED).Where("perm_code > 0")
	query.Cols("resource_args").Find(&limitedWhiteList)
	loadedWhiteList = true
	return limitedWhiteList
}

// 获取正常用户权限可访问的网址
func (a UserAuth) GetRegularPermissions(roles []string) (perms []auth.IPermission) {
	var objs []*Permission
	QueryPermissions(roles...).Find(&objs)
	for _, obj := range objs {
		perms = append(perms, obj)
	}
	return
}

// 获取超级用户权限可访问的网址，不再检查正常用户权限
func (a UserAuth) GetSuperPermissions(roles []string) (perms []auth.IPermission) {
	var objs []*Permission
	QueryPermissions(ROLE_NAME_SUPER).Find(&objs)
	for _, obj := range objs {
		perms = append(perms, obj)
	}
	return
}

func QueryPermissions(roles ...string) *xorm.Session {
	query := db.Table(Permission{}).Where("revoked_at IS NULL")
	query = query.In("resource_type", "menu", "url")
	if len(roles) > 1 {
		query = query.In("role_name", roles)
	} else {
		roles = append(roles, "")
		query = query.Where("role_name = ?", roles[0])
	}
	return query
}
