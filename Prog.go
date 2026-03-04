package big

import (
	"fmt"
	"math/big"
	"math/bits"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Prog is a compiled expression that can be executed many times with different
// parameter values.
type Prog struct {
	ops         []tOp
	paramsCount int
	stackSize   int
	stackPool   sync.Pool
	opsBuf      [32]tOp
}

// Init compiles exprAsString into p. paramNames defines allowed parameter names
// and their positional order for [Prog.Exec]. Init panics on invalid syntax,
// unknown functions, or unknown parameters.
func (p *Prog) Init(exprAsString string, paramNames ...string) {
	tokens := make([]tToken, 0, 32)
	tokens = tokenize(exprAsString, tokens)
	rpn := make([]tToken, 0, 32)
	rpn = toRPN(tokens, rpn)
	ops := p.opsBuf[:0]
	paramNameToIdx := make(map[string]int)
	for i, paramName := range paramNames {
		paramNameToIdx[paramName] = i
	}
	ops, stackSize := compileOps(rpn, paramNameToIdx, ops)
	p.ops = ops
	p.paramsCount = len(paramNames)
	p.stackSize = stackSize
}

func compileOps(rpnTokens []tToken, paramNameToIdx map[string]int, ops []tOp) ([]tOp, int) {
	depth := 0
	maxDepth := 0
	for _, token := range rpnTokens {
		switch token.tokenType {
		case tokenTypeNum:
			n := new(big.Int)
			if _, ok := n.SetString(token.val, 10); !ok {
				panic(fmt.Sprintf("invalid number: %s", token.val))
			}
			ops = append(ops, tOp{
				Cat:   opCategoryLoad,
				Const: n,
			})
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		case tokenTypeParam:
			idx, ok := paramNameToIdx[token.val]
			if !ok {
				panic(fmt.Sprintf("undefined variable: %s", token.val))
			}
			ops = append(ops, tOp{
				Cat:      opCategoryLoad,
				paramIdx: idx,
			})
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		case tokenTypeBinaryOp:
			switch token.oper {
			case operAdd:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Add})
			case operSub:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Sub})
			case operMul:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Mul})
			case operDiv:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Div})
			case operMod:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Mod})
			case operLsh:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: lshOp})
			case operRsh:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: rshOp})
			case operAnd:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).And})
			case operOr:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Or})
			case operXor:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).Xor})
			case operAndNot:
				ops = append(ops, tOp{Cat: opCategoryBinary, Fn: (*big.Int).AndNot})
			default:
				msg := fmt.Sprintf("unknown operator code: %d", token.oper)
				panic(msg)
			}
			depth--
		case tokenTypeUnaryOp:
			switch token.oper {
			case operNeg:
				ops = append(ops, tOp{Cat: opCategoryUnary, Fn: func(a, _, _ *big.Int) *big.Int { return a.Neg(a) }})
			case operNot:
				ops = append(ops, tOp{Cat: opCategoryUnary, Fn: func(a, _, _ *big.Int) *big.Int { return a.Not(a) }})
			default:
				msg := fmt.Sprintf("unknown unary operator code: %d", token.oper)
				panic(msg)
			}
		case tokenTypeFn:
			fn, cat := lookupFn(token.val)
			if fn == nil {
				msg := fmt.Sprintf("unknown function: %s", token.val)
				panic(msg)
			}
			ops = append(ops, tOp{Cat: cat, Fn: fn})
			if cat == opCategoryBinary {
				depth--
			}
		default:
			panic("unhandled token")
		}
	}
	return ops, maxDepth
}

func toRPN(tokens []tToken, output []tToken) []tToken {
	var stack []tToken
	for _, token := range tokens {
		switch token.tokenType {
		case tokenTypeNum, tokenTypeParam:
			output = append(output, token)
			for len(stack) > 0 && stack[len(stack)-1].tokenType == tokenTypeUnaryOp {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
		case tokenTypeFn:
			stack = append(stack, token)
		case tokenTypeUnaryOp:
			stack = append(stack, token)
		case tokenTypeComma:
			found := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.tokenType == tokenTypeLParen {
					found = true
					break
				}
				output = append(output, top)
				stack = stack[:len(stack)-1]
			}
			if !found {
				panic("misplaced function argument separator")
			}
		case tokenTypeBinaryOp:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.tokenType == tokenTypeUnaryOp {
					output = append(output, top)
					stack = stack[:len(stack)-1]
					continue
				}
				if top.tokenType == tokenTypeBinaryOp && precedence(token.oper) <= precedence(top.oper) {
					output = append(output, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, token)
		case tokenTypeLParen:
			stack = append(stack, token)
		case tokenTypeRParen:
			found := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.tokenType == tokenTypeLParen {
					found = true
					break
				}
				output = append(output, top)
			}
			if !found {
				panic("mismatched parentheses")
			}
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.tokenType != tokenTypeFn && top.tokenType != tokenTypeUnaryOp {
					break
				}
				output = append(output, top)
				stack = stack[:len(stack)-1]
			}
		default:
			panic("unhandled token")
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.tokenType == tokenTypeLParen {
			panic("mismatched parentheses")
		}
		output = append(output, top)
	}
	return output
}

func precedence(op tOper) int {
	switch op {
	case operAdd, operSub, operOr, operXor:
		return 1
	case operMul, operDiv, operMod, operLsh, operRsh, operAnd, operAndNot:
		return 2
	default:
		return 0
	}
}

func shiftAmount(v *big.Int) uint {
	if v.Sign() < 0 {
		panic("negative shift count")
	}
	if v.BitLen() > bits.UintSize {
		panic("shift count overflows uint")
	}
	return uint(v.Uint64())
}

func lshOp(z, x, y *big.Int) *big.Int {
	return z.Lsh(x, shiftAmount(y))
}

func rshOp(z, x, y *big.Int) *big.Int {
	return z.Rsh(x, shiftAmount(y))
}

func lookupFn(name string) (tOpFn, tOpCat) {
	switch name {
	case "sqrt":
		return func(a, _, _ *big.Int) *big.Int { return a.Sqrt(a) }, opCategoryUnary
	case "abs":
		return func(a, _, _ *big.Int) *big.Int { return a.Abs(a) }, opCategoryUnary
	case "modInverse":
		return (*big.Int).ModInverse, opCategoryBinary
	case "modSqrt":
		return (*big.Int).ModSqrt, opCategoryBinary
	case "quo":
		return (*big.Int).Quo, opCategoryBinary
	case "rem":
		return (*big.Int).Rem, opCategoryBinary
	default:
		return nil, 0
	}
}

func tokenize(expr string, tokens []tToken) []tToken {
	var prevTyp tTokenType
	for i := 0; i < len(expr); {
		r, size := utf8.DecodeRuneInString(expr[i:])
		if r == utf8.RuneError && size == 1 {
			panic("invalid utf-8 encoding")
		}
		if unicode.IsSpace(r) {
			i += size
			continue
		}
		var tokenType tTokenType
		var val string
		var oper tOper
		switch r {
		case '+':
			tokenType = tokenTypeBinaryOp
			oper = operAdd
			i += size
		case '*':
			tokenType = tokenTypeBinaryOp
			oper = operMul
			i += size
		case '<':
			tokenType = tokenTypeBinaryOp
			oper, i = tokenizeShiftOp(expr, i, size, r)
		case '>':
			tokenType = tokenTypeBinaryOp
			oper, i = tokenizeShiftOp(expr, i, size, r)
		case '/':
			tokenType = tokenTypeBinaryOp
			oper = operDiv
			i += size
		case '%':
			tokenType = tokenTypeBinaryOp
			oper = operMod
			i += size
		case '|':
			tokenType = tokenTypeBinaryOp
			oper = operOr
			i += size
		case '^':
			tokenType = tokenTypeBinaryOp
			oper = operXor
			i += size
		case '-':
			if isUnaryContext(len(tokens), prevTyp) {
				tokenType = tokenTypeUnaryOp
				oper = operNeg
			} else {
				tokenType = tokenTypeBinaryOp
				oper = operSub
			}
			i += size
		case '~':
			if isUnaryContext(len(tokens), prevTyp) {
				tokenType = tokenTypeUnaryOp
				oper = operNot
			} else {
				tokenType = tokenTypeBinaryOp
			}
			i += size
		case '(':
			tokenType = tokenTypeLParen
			i += size
		case ',':
			tokenType = tokenTypeComma
			i += size
		case ')':
			tokenType = tokenTypeRParen
			i += size
		case '&':
			tokenType = tokenTypeBinaryOp
			oper = operAnd
			i += size
			if i < len(expr) {
				next, nextSize := utf8.DecodeRuneInString(expr[i:])
				if next == utf8.RuneError && nextSize == 1 {
					panic("invalid utf-8 encoding")
				}
				if next == '^' {
					oper = operAndNot
					i += nextSize
				}
			}
		default:
			switch {
			case unicode.IsDigit(r):
				start := i
				i += size
				for i < len(expr) {
					next, nextSize := utf8.DecodeRuneInString(expr[i:])
					if next == utf8.RuneError && nextSize == 1 {
						panic("invalid utf-8 encoding")
					}
					if !unicode.IsDigit(next) {
						break
					}
					i += nextSize
				}
				tokenType = tokenTypeNum
				val = expr[start:i]
			case unicode.IsLetter(r):
				start := i
				i += size
				for i < len(expr) {
					next, nextSize := utf8.DecodeRuneInString(expr[i:])
					if next == utf8.RuneError && nextSize == 1 {
						panic("invalid utf-8 encoding")
					}
					if !unicode.IsLetter(next) && !unicode.IsDigit(next) {
						break
					}
					i += nextSize
				}
				val = expr[start:i]
				if nextNonSpaceIsLParen(expr, i) {
					tokenType = tokenTypeFn
				} else {
					tokenType = tokenTypeParam
				}
			default:
				panic(fmt.Sprintf("invalid character: %c", r))
			}
		}
		tokens = append(tokens, tToken{tokenType: tokenType, val: val, oper: oper})
		prevTyp = tokenType
	}
	return tokens
}

func tokenizeShiftOp(expr string, i int, size int, r rune) (tOper, int) {
	i += size
	if i >= len(expr) {
		msg := fmt.Sprintf("invalid character: %c", r)
		panic(msg)
	}
	next, nextSize := utf8.DecodeRuneInString(expr[i:])
	if next == utf8.RuneError && nextSize == 1 {
		panic("invalid utf-8 encoding")
	}
	if next != r {
		msg := fmt.Sprintf("invalid character: %c", r)
		panic(msg)
	}
	i += nextSize
	switch r {
	case '<':
		return operLsh, i
	case '>':
		return operRsh, i
	default:
		msg := fmt.Sprintf("invalid character: %c", r)
		panic(msg)
	}
}

func nextNonSpaceIsLParen(expr string, i int) bool {
	for i < len(expr) {
		r, size := utf8.DecodeRuneInString(expr[i:])
		if r == utf8.RuneError && size == 1 {
			panic("invalid utf-8 encoding")
		}
		if !unicode.IsSpace(r) {
			return r == '('
		}
		i += size
	}
	return false
}

func isUnaryContext(tokensLen int, prevTyp tTokenType) bool {
	if tokensLen == 0 {
		return true
	}
	switch prevTyp {
	case tokenTypeLParen, tokenTypeBinaryOp, tokenTypeUnaryOp:
		return true
	default:
		return false
	}
}

// Exec runs the compiled program with paramVals and returns the computed value.
// The number and order of values must match paramNames passed to [Compile] or
// [Prog.Init]. Exec is safe for concurrent use by multiple goroutines.
// Exec panics when the argument count does not match.
func (p *Prog) Exec(paramVals ...*big.Int) *big.Int {
	if len(paramVals) != p.paramsCount {
		msg := fmt.Sprintf("expected %d param values, got %d", p.paramsCount, len(paramVals))
		panic(msg)
	}
	var stack []big.Int
	stackAsAny := p.stackPool.Get()
	if stackAsAny != nil {
		stack = stackAsAny.([]big.Int)
	} else {
		stack = make([]big.Int, p.stackSize)
	}
	defer p.stackPool.Put(stack)
	top := 0
	for _, op := range p.ops {
		switch op.Cat {
		case opCategoryLoad:
			a := &stack[top]
			var v *big.Int
			if c := op.Const; c != nil {
				v = c
			} else {
				v = paramVals[op.paramIdx]
			}
			a.Set(v)
			top++
		case opCategoryUnary:
			a := &stack[top-1]
			op.Fn(a, nil, nil)
		case opCategoryBinary:
			top--
			a := &stack[top-1]
			b := &stack[top]
			op.Fn(a, a, b)
		}
	}
	return new(big.Int).Set(&stack[0])
}
