package access

import (
	"sort"

	"github.com/k0kubun/pp"
)

// 操作列表
const (
	VIEW uint16 = 2 << iota
	DISABLE
	REMOVE
	EDIT
	CREATE
	GET
	POST
	GRANT
	ALL         = ^uint16(0) // 65535
	NONE uint16 = 0          // 无权限
)

var verbose = false

func init() {
	if !verbose {
		return
	}
	// 测试操作列表具体数值
	var codes []int
	for c := range AccessTitles {
		codes = append(codes, int(c))
	}
	sort.Ints(codes)
	for _, cc := range codes {
		c := uint16(cc)
		pp.Println(c, cc, AccessTitles[c])
	}
}

var (
	AccessNames = map[uint16]string{
		VIEW: "view", DISABLE: "disable", REMOVE: "remove",
		EDIT: "edit", CREATE: "create", GET: "get", POST: "post",
		GRANT: "grant", ALL: "all", NONE: "",
	}
	AccessTitles = map[uint16]string{
		VIEW: "查看", DISABLE: "禁用", REMOVE: "删除",
		EDIT: "编辑", CREATE: "新建", GET: "GET", POST: "POST",
		GRANT: "授权", ALL: "全部", NONE: "无",
	}
)

func ContainAction(perm, act uint16, strict bool) bool {
	if strict && act == NONE {
		return false
	}
	return perm == ALL || perm&act == act
}

// 分解出具体权限
func ParsePermNames(perm uint16) (codes []uint16, names []string) {
	if perm == NONE {
		return
	} else if perm == ALL {
		codes = append(codes, ALL)
		names = append(names, AccessNames[ALL])
		return
	}
	for code, name := range AccessNames {
		if ContainAction(perm, code, true) {
			codes = append(codes, code)
			names = append(names, name)
		}
	}
	return
}

// 找出权限的中文名称
func GetPermTitles(codes []uint16) (titles []string) {
	title, ok := "", false
	for _, code := range codes {
		if title, ok = AccessTitles[code]; !ok {
			title = "未知"
		}
		titles = append(titles, title)
	}
	return
}
