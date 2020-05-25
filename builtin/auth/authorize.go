package auth

import (
	"fmt"

	"gitee.com/azhai/xorm-refactor/builtin/usertype"
	"github.com/azhai/gozzo-utils/common"
)

type IPermission interface {
	CheckPerm(act uint16, url string) bool
}

type IUserAuth interface {
	// 用户分类，无法区分内部用户和普通用户
	GetUserType() (utype usertype.UserType, err error)

	// 用户拥有的角色
	GetUserRoles() (roles []string, err error)

	// 是否静态资源网址
	IsStaticResourceUrl(url string) bool

	// 获取可公开访问的网址
	GetAnonymousOpenUrls() (urls []string)

	// 获取受限用户黑名单中的的网址，与白名单二选一
	GetLimitedBlackListUrls() (urls []string)

	// 获取受限用户白名单中的的网址，不再检查正常用户权限，与黑名单二选一
	GetLimitedWhiteListUrls() (urls []string)

	// 获取正常用户权限可访问的网址
	GetRegularPermissions(roles []string) (perms []IPermission)

	// 获取超级用户权限可访问的网址，不再检查正常用户权限
	GetSuperPermissions(roles []string) (perms []IPermission)
}

// 用户鉴权
func Authorize(auth IUserAuth, act uint16, url string) error {
	var err error

	// 1. 静态资源，直接放行
	if auth.IsStaticResourceUrl(url) {
		return nil
	}

	var utype usertype.UserType
	if utype, err = auth.GetUserType(); err != nil { // 出错了
		return err
	}

	// 2. 匿名用户，如果是公开资源放行，否则失败
	if utype == usertype.ANONYMOUS || utype == usertype.FORBIDDEN {
		if urls := auth.GetAnonymousOpenUrls(); len(urls) > 0 {
			if !common.StartStringList(url, urls) {
				err = fmt.Errorf("已注册用户可访问，请您先登录！")
			}
		}
		return err // 匿名用户到此为止
	}

	// 3. 受限用户，优先判断黑名单，此网址在黑名单中则失败
	if utype == usertype.LIMITED {
		if urls := auth.GetLimitedBlackListUrls(); len(urls) > 0 { // 二选一
			if common.StartStringList(url, urls) {
				err = fmt.Errorf("您的账号无权限访问，请联系客服！")
				return err
			}
		} else if urls := auth.GetLimitedWhiteListUrls(); len(urls) > 0 { // 二选一
			if common.StartStringList(url, urls) {
				return nil
			}
		}
	}

	var roles []string
	if roles, err = auth.GetUserRoles(); err != nil { // 出错了
		return err
	}

	// 4. 超级用户，如果有此权限则放行
	if utype == usertype.SUPER {
		if perms := auth.GetSuperPermissions(roles); len(perms) > 0 {
			for _, perm := range perms {
				if perm.CheckPerm(act, url) {
					return nil
				}
			}
		}
	}

	// 5. 正常用户，如果有此权限则放行，内容最多，放在最后
	if perms := auth.GetRegularPermissions(roles); len(perms) > 0 {
		for _, perm := range perms {
			if perm.CheckPerm(act, url) {
				return nil
			}
		}
	}

	// 6. 权限不明确网址，失败
	err = fmt.Errorf("找不到此网址，请核实后访问！")
	return err
}
