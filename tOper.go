package big

type tOper int

const (
	operNone tOper = iota
	operAdd
	operSub
	operMul
	operDiv
	operMod
	operLsh
	operRsh
	operAnd
	operOr
	operXor
	operAndNot
	operNeg
	operNot
)
