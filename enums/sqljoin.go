//go:generate enumer -type=SqlJoin -transform=BLANK -text
package enums

import "strings"

type SqlJoin int

// 数据表联接
const (
	Join SqlJoin = iota
	InnerJoin
	OuterJoin
	CrossJoin

	LeftJoin SqlJoin = iota
	LeftInnerJoin
	LeftOuterJoin

	RightJoin SqlJoin = iota + 1
	RightInnerJoin
	RightOuterJoin
)

func (i SqlJoin) Subject() string {
	return strings.TrimSuffix(i.String(), " JOIN")
}
