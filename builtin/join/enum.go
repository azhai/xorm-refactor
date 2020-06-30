//go:generate enumer -type=JoinOp -transform=BLANK -text
package join

type JoinOp int

const (
	Join JoinOp = iota
	InnerJoin
	OuterJoin
	CrossJoin
	LeftJoin       = 4
	RightJoin      = 8
	LeftInnerJoin  = LeftJoin + InnerJoin
	LeftOuterJoin  = LeftJoin + OuterJoin
	RightInnerJoin = RightJoin + InnerJoin
	RightOuterJoin = RightJoin + OuterJoin
)
