package usertype

type UserType int

// 用户分类
const (
	ANONYMOUS UserType = iota // 匿名用户（未登录/未注册）
	FORBIDDEN                 // 封禁用户（有违规被封号）
	LIMITED                   // 受限用户（未过审或被降级）
	REGULAR                   // 正常用户（正式会员）
	SUPER                     // 超级用户（后台管理权限）
)
