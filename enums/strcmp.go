//go:generate enumer -type=StrCmp -text
package enums

type StrCmp int

// 字符串比较方式
const (
	Omit            StrCmp = iota // 不比较
	Contains                      // 包含
	StartsWith                    // 打头
	EndsWith                      // 结尾
	CaseInsensitive               // 不分大小写
	IgnoreSpaces                  // 忽略空格
	Equal                         // 相等
)
