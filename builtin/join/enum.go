package join

type JoinOp string

const (
	JOIN             JoinOp = ""
	INNER_JOIN       JoinOp = "INNER"
	OUTER_JOIN       JoinOp = "OUTER"
	CROSS_JOIN       JoinOp = "CROSS"
	LEFT_JOIN        JoinOp = "LEFT"
	RIGHT_JOIN       JoinOp = "RIGHT"
	LEFT_INNER_JOIN         = LEFT_JOIN + " " + INNER_JOIN
	LEFT_OUTER_JOIN         = LEFT_JOIN + " " + OUTER_JOIN
	RIGHT_INNER_JOIN        = RIGHT_JOIN + " " + INNER_JOIN
	RIGHT_OUTER_JOIN        = RIGHT_JOIN + " " + OUTER_JOIN
)
