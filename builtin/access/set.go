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
	EXPORT
	IMPORT
	BATCH
	GET
	POST
	PUT
	DELETE
	OPTION
	GRANT
	REVOKE
	ALL         = ^uint16(0) // 65535
	NONE uint16 = 0          // 无权限
)

var verbose = false

func init() {
	for c := range AccessTitles {
		AccessCodes = append(AccessCodes, c)
	}
	sort.Sort(CodeSlice(AccessCodes))
	if !verbose {
		return
	}
	// 测试操作列表具体数值
	for _, c := range AccessCodes {
		pp.Println(c, AccessTitles[c])
	}
}

var (
	AccessCodes []uint16
	AccessNames = map[uint16]string{
		VIEW: "view", DISABLE: "disable", REMOVE: "remove", EDIT: "edit",
		CREATE: "create", EXPORT: "export", IMPORT: "import", BATCH: "batch",
		GET: "get", POST: "post", PUT: "put", DELETE: "delete", OPTION: "option",
		GRANT: "grant", REVOKE: "revoke", ALL: "all", NONE: "",
	}
	AccessTitles = map[uint16]string{
		VIEW: "查看", DISABLE: "禁用", REMOVE: "删除", EDIT: "编辑",
		CREATE: "新建", EXPORT: "导出", IMPORT: "导入", BATCH: "批量",
		GET: "GET", POST: "POST", PUT: "PUT", DELETE: "DELETE", OPTION: "OPTION",
		GRANT: "授权", REVOKE: "撤销", ALL: "全部", NONE: "无",
	}
)

type CodeSlice []uint16

func (p CodeSlice) Len() int           { return len(p) }
func (p CodeSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p CodeSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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
	for _, c := range AccessCodes {
		if ContainAction(perm, c, true) {
			codes = append(codes, c)
			names = append(names, AccessNames[c])
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
