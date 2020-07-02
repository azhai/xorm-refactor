//go:generate enumer -type=UserType -text
package enums

type UserType int

// 用户分类
const (
	Anonymous UserType = iota // 匿名用户（未登录/未注册）
	Forbidden                 // 封禁用户（有违规被封号）
	Limited                   // 受限用户（未过审或被降级）
	Regular                   // 正常用户（正式会员）
	Super                     // 超级用户（后台管理权限）
)
