package enums

import (
	"sort"
	"strings"

	"gitee.com/azhai/xorm-refactor/utils"
)

// 比较是否相符
func MatchString(a, b string, cmp StrCmp) bool {
	switch cmp {
	case Omit:
		return true
	case Contains:
		return strings.Contains(a, b)
	case StartsWith:
		return strings.HasPrefix(a, b)
	case EndsWith:
		return strings.HasSuffix(a, b)
	case CaseInsensitive:
		return strings.EqualFold(a, b)
	case IgnoreSpaces:
		a, b = utils.RemoveSpaces(a), utils.RemoveSpaces(b)
		return strings.EqualFold(a, b)
	default: // 包括 strcmp.EQUAL
		return strings.Compare(a, b) == 0
	}
}

// 是否在字符串列表中，主要适用于 strcmp.Contains 和 strcmp.EndsWith
func MatchStringList(x string, lst []string, cmp StrCmp) bool {
	for _, y := range lst {
		if MatchString(x, y, cmp) {
			return true
		}
	}
	return false
}

// 是否在字符串列表中，仅适用于 strcmp.Equal 和 strcmp.StartsWith
func SearchStringList(x string, lst []string, cmp StrCmp) bool {
	if !sort.StringsAreSorted(lst) {
		sort.Strings(lst)
	}
	if i := sort.SearchStrings(lst, x); i < len(lst) {
		return MatchString(x, lst[i], cmp)
	}
	return false
}

// 是否在字符串列表中
func CompareStringList(x string, lst []string, cmp StrCmp) bool {
	if size := len(lst); size == 0 {
		return false
	}
	switch cmp {
	case Omit:
		return true
	case Contains, EndsWith:
		return MatchStringList(x, lst, cmp)
	case CaseInsensitive:
		x = strings.ToLower(x)
		for i, y := range lst {
			lst[i] = strings.ToLower(y)
		}
	case IgnoreSpaces:
		x = utils.RemoveSpaces(x)
		for i, y := range lst {
			lst[i] = utils.RemoveSpaces(y)
		}
	}
	return SearchStringList(x, lst, cmp)
}

// 是否在字符串列表中
func InStringList(x string, lst []string) bool {
	return len(lst) > 0 && SearchStringList(x, lst, Equal)
}

// 是否在字符串列表中，比较方式是有任何一个开头符合
func StartStringList(x string, lst []string) bool {
	return len(lst) > 0 && SearchStringList(x, lst, StartsWith)
}

// 是否在字符串列表中，比较方式是有任何一个开头符合
func EndStringList(x string, lst []string) bool {
	return len(lst) > 0 && MatchStringList(x, lst, EndsWith)
}

// 比较两个列表的长度
func CompareSubsetLength(lst1, lst2 []string, strict bool) bool {
	if len(lst1) > len(lst2) {
		return false
	}
	if strict && len(lst1) == len(lst2) {
		return false
	}
	return true
}

// lst1 是否 lst2 的（真）子集
func CompareSubsetList(lst1, lst2 []string, cmp StrCmp, strict bool) bool {
	if !CompareSubsetLength(lst1, lst2, strict) {
		return false
	}
	for _, x := range lst1 {
		if !CompareStringList(x, lst2, cmp) {
			return false
		}
	}
	return true
}

// lst1 是否 lst2 的（真）子集
func IsSubsetList(lst1, lst2 []string, strict bool) bool {
	if !CompareSubsetLength(lst1, lst2, strict) {
		return false
	}
	for _, x := range lst1 {
		if !InStringList(x, lst2) {
			return false
		}
	}
	return true
}
