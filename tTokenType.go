package bigint

type tTokenType int

const (
	tokenTypeNum tTokenType = iota
	tokenTypeParam
	tokenTypeBinaryOp
	tokenTypeUnaryOp
	tokenTypeFn
	tokenTypeComma
	tokenTypeLParen
	tokenTypeRParen
)
