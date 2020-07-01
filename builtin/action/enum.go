//go:generate enumer -type=Action -text
package action

type Action int

// 操作列表
const (
	Noop Action = 0 // 无操作
	View Action = 2 << iota
	Draft
	Delete
	Add
	Edit
	Export
	Batch
)
