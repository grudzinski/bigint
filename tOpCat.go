package bigint

type tOpCat int

const (
	opCategoryLoad tOpCat = iota
	opCategoryUnary
	opCategoryBinary
)
