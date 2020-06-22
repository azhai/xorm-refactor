package utils

import (
	"sort"
	"strings"
)

// 字符串比较方式
const (
	CMP_STRING_OMIT             = iota // 不比较
	CMP_STRING_CONTAINS                // 包含
	CMP_STRING_STARTSWITH              // 打头
	CMP_STRING_ENDSWITH                // 结尾
	CMP_STRING_CASE_INSENSITIVE        // 不分大小写
	CMP_STRING_IGNORE_SPACES           // 忽略空格
	CMP_STRING_EQUAL                   // 相等
)

// 比较是否相符
func MatchString(a, b string, cmp int) bool {
	switch cmp {
	case CMP_STRING_OMIT:
		return true
	case CMP_STRING_CONTAINS:
		return strings.Contains(a, b)
	case CMP_STRING_STARTSWITH:
		return strings.HasPrefix(a, b)
	case CMP_STRING_ENDSWITH:
		return strings.HasSuffix(a, b)
	case CMP_STRING_CASE_INSENSITIVE:
		return strings.EqualFold(a, b)
	case CMP_STRING_IGNORE_SPACES:
		a, b = RemoveSpaces(a), RemoveSpaces(b)
		return strings.EqualFold(a, b)
	default: // 包括 CMP_STRING_EQUAL
		return strings.Compare(a, b) == 0
	}
}

// 是否在字符串列表中，主要适用于 CMP_STRING_CONTAINS 和 CMP_STRING_ENDSWITH
func MatchStringList(x string, lst []string, cmp int) bool {
	for _, y := range lst {
		if MatchString(x, y, cmp) {
			return true
		}
	}
	return false
}

// 是否在字符串列表中，仅适用于 CMP_STRING_EQUAL 和 CMP_STRING_STARTSWITH
func SearchStringList(x string, lst []string, cmp int) bool {
	if !sort.StringsAreSorted(lst) {
		sort.Strings(lst)
	}
	if i := sort.SearchStrings(lst, x); i < len(lst) {
		return MatchString(x, lst[i], cmp)
	}
	return false
}

// 是否在字符串列表中
func CompareStringList(x string, lst []string, cmp int) bool {
	if size := len(lst); size == 0 {
		return false
	}
	switch cmp {
	case CMP_STRING_OMIT:
		return true
	case CMP_STRING_CONTAINS, CMP_STRING_ENDSWITH:
		return MatchStringList(x, lst, cmp)
	case CMP_STRING_CASE_INSENSITIVE:
		x = strings.ToLower(x)
		for i, y := range lst {
			lst[i] = strings.ToLower(y)
		}
	case CMP_STRING_IGNORE_SPACES:
		x = RemoveSpaces(x)
		for i, y := range lst {
			lst[i] = RemoveSpaces(y)
		}
	}
	return SearchStringList(x, lst, cmp)
}

// 是否在字符串列表中
func InStringList(x string, lst []string) bool {
	return len(lst) > 0 && SearchStringList(x, lst, CMP_STRING_EQUAL)
}

// 是否在字符串列表中，比较方式是有任何一个开头符合
func StartStringList(x string, lst []string) bool {
	return len(lst) > 0 && SearchStringList(x, lst, CMP_STRING_STARTSWITH)
}

// 是否在字符串列表中，比较方式是有任何一个开头符合
func EndStringList(x string, lst []string) bool {
	return len(lst) > 0 && MatchStringList(x, lst, CMP_STRING_ENDSWITH)
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
func CompareSubsetList(lst1, lst2 []string, cmp int, strict bool) bool {
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
