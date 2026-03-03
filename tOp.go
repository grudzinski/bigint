package big

import "math/big"

type tOp struct {
	Cat      tOpCat
	Fn       tOpFn
	paramIdx int
	Const    *big.Int
}
