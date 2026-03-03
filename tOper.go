package big

type tOper int

const (
	operNone tOper = iota
	operAdd
	operSub
	operMul
	operDiv
	operMod
	operAnd
	operOr
	operXor
	operAndNot
	operNeg
	operNot
)
