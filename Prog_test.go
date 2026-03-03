package big

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
)

var evalComplexRawPool = sync.Pool{
	New: func() any {
		return new([3]big.Int)
	},
}

func TestCompileAndEval(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		vars     []string
		values   []*big.Int
		expected *big.Int
		wantErr  bool
	}{
		{
			name:     "simple multiplication and division",
			expr:     "(a * b) / c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(5)},
			expected: big.NewInt(40),
		},
		{
			name:     "addition and subtraction",
			expr:     "(a + b) - c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(100), big.NewInt(50), big.NewInt(30)},
			expected: big.NewInt(120),
		},
		{
			name:     "modulo operation",
			expr:     "a % b",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(17), big.NewInt(5)},
			expected: big.NewInt(2),
		},
		{
			name:     "complex expression",
			expr:     "(a + b) * (c - d)",
			vars:     []string{"a", "b", "c", "d"},
			values:   []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30), big.NewInt(5)},
			expected: big.NewInt(750),
		},
		{
			name:     "with numeric constant",
			expr:     "(x * 100) / y",
			vars:     []string{"x", "y"},
			values:   []*big.Int{big.NewInt(50), big.NewInt(10)},
			expected: big.NewInt(500),
		},
		{
			name:     "sqrt",
			expr:     "sqrt(a)",
			vars:     []string{"a"},
			values:   []*big.Int{big.NewInt(16)},
			expected: big.NewInt(4),
		},
		{
			name:     "abs",
			expr:     "abs(a)",
			vars:     []string{"a"},
			values:   []*big.Int{big.NewInt(-5)},
			expected: big.NewInt(5),
		},
		{
			name:    "undefined variable",
			expr:    "a * b",
			vars:    []string{"a"},
			values:  []*big.Int{big.NewInt(10)},
			wantErr: true,
		},
		{
			name:    "mismatched parentheses",
			expr:    "(a * b",
			vars:    []string{"a", "b"},
			values:  []*big.Int{big.NewInt(10), big.NewInt(20)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("Compile() panicked = %v, wantErr %v", r, tt.wantErr)
					}
				} else if tt.wantErr {
					t.Errorf("Compile() did not panic, wantErr %v", tt.wantErr)
				}
			}()

			p := Compile(tt.expr, tt.vars...)

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("Exec() panicked = %v, wantErr %v", r, tt.wantErr)
					}
				} else if tt.wantErr {
					t.Errorf("Exec() did not panic, wantErr %v", tt.wantErr)
				}
			}()

			result := p.Exec(tt.values...)

			if !tt.wantErr && result.Cmp(tt.expected) != 0 {
				t.Errorf("Exec() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		vars     []string
		values   []*big.Int
		expected *big.Int
	}{
		{
			name:     "multiplication before addition",
			expr:     "a + b * c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(2), big.NewInt(3), big.NewInt(4)},
			expected: big.NewInt(14), // 2 + (3 * 4) = 14
		},
		{
			name:     "parentheses override precedence",
			expr:     "(a + b) * c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(2), big.NewInt(3), big.NewInt(4)},
			expected: big.NewInt(20), // (2 + 3) * 4 = 20
		},
		{
			name:     "and before add",
			expr:     "a + b & c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(8), big.NewInt(6), big.NewInt(3)},
			expected: big.NewInt(10), // 8 + (6 & 3) = 10
		},
		{
			name:     "and-not before xor",
			expr:     "a ^ b &^ c",
			vars:     []string{"a", "b", "c"},
			values:   []*big.Int{big.NewInt(12), big.NewInt(14), big.NewInt(5)},
			expected: big.NewInt(6), // 12 ^ (14 &^ 5) = 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Compile(tt.expr, tt.vars...)

			result := p.Exec(tt.values...)

			if result.Cmp(tt.expected) != 0 {
				t.Errorf("Exec() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBitwiseOperators(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		vars     []string
		values   []*big.Int
		expected *big.Int
	}{
		{
			name:     "bitwise and",
			expr:     "a & b",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(6), big.NewInt(3)},
			expected: big.NewInt(2),
		},
		{
			name:     "bitwise or",
			expr:     "a | b",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(6), big.NewInt(3)},
			expected: big.NewInt(7),
		},
		{
			name:     "bitwise xor",
			expr:     "a ^ b",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(6), big.NewInt(3)},
			expected: big.NewInt(5),
		},
		{
			name:     "bitwise and not",
			expr:     "a &^ b",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(6), big.NewInt(3)},
			expected: big.NewInt(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Compile(tt.expr, tt.vars...)
			got := p.Exec(tt.values...)
			if got.Cmp(tt.expected) != 0 {
				t.Fatalf("Exec() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSupportedOperatorsAndFunctions(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		vars     []string
		values   []*big.Int
		expected *big.Int
	}{
		{name: "operator add", expr: "a + b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(7), big.NewInt(5)}, expected: big.NewInt(12)},
		{name: "operator sub", expr: "a - b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(7), big.NewInt(5)}, expected: big.NewInt(2)},
		{name: "operator mul", expr: "a * b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(7), big.NewInt(5)}, expected: big.NewInt(35)},
		{name: "operator div", expr: "a / b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(20), big.NewInt(6)}, expected: big.NewInt(3)},
		{name: "operator mod", expr: "a % b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(20), big.NewInt(6)}, expected: big.NewInt(2)},
		{name: "operator and", expr: "a & b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(14), big.NewInt(11)}, expected: big.NewInt(10)},
		{name: "operator or", expr: "a | b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(10), big.NewInt(5)}, expected: big.NewInt(15)},
		{name: "operator xor", expr: "a ^ b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(10), big.NewInt(5)}, expected: big.NewInt(15)},
		{name: "operator and not", expr: "a &^ b", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(14), big.NewInt(5)}, expected: big.NewInt(10)},
		{name: "operator unary neg", expr: "-a", vars: []string{"a"}, values: []*big.Int{big.NewInt(7)}, expected: big.NewInt(-7)},
		{name: "operator unary not", expr: "~a", vars: []string{"a"}, values: []*big.Int{big.NewInt(7)}, expected: big.NewInt(-8)},
		{name: "function sqrt", expr: "sqrt(a)", vars: []string{"a"}, values: []*big.Int{big.NewInt(81)}, expected: big.NewInt(9)},
		{name: "function abs", expr: "abs(a)", vars: []string{"a"}, values: []*big.Int{big.NewInt(-9)}, expected: big.NewInt(9)},
		{name: "function modInverse", expr: "modInverse(a, m)", vars: []string{"a", "m"}, values: []*big.Int{big.NewInt(3), big.NewInt(11)}, expected: big.NewInt(4)},
		{name: "function quo", expr: "quo(a, b)", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(20), big.NewInt(6)}, expected: big.NewInt(3)},
		{name: "function rem", expr: "rem(a, b)", vars: []string{"a", "b"}, values: []*big.Int{big.NewInt(20), big.NewInt(6)}, expected: big.NewInt(2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Compile(tt.expr, tt.vars...)
			got := p.Exec(tt.values...)
			if got.Cmp(tt.expected) != 0 {
				t.Fatalf("Exec(%q) = %v, want %v", tt.expr, got, tt.expected)
			}
		})
	}

	t.Run("function modSqrt", func(t *testing.T) {
		a := big.NewInt(4)
		modulus := big.NewInt(7)
		p := Compile("modSqrt(a, p)", "a", "p")
		root := p.Exec(a, modulus)

		// Validate the result by property to avoid depending on a specific root.
		got := new(big.Int).Mul(root, root)
		got.Mod(got, modulus)
		want := new(big.Int).Mod(a, modulus)
		if got.Cmp(want) != 0 {
			t.Fatalf("modSqrt property failed: (%v^2) %% %v = %v, want %v", root, modulus, got, want)
		}
	})
}

func TestCompileOpsUnaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		oper     tOper
		input    int64
		expected int64
	}{
		{
			name:     "neg",
			oper:     operNeg,
			input:    2,
			expected: -2,
		},
		{
			name:     "not",
			oper:     operNot,
			input:    2,
			expected: -3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpn := []tToken{
				{tokenType: tokenTypeNum, val: "2"},
				{tokenType: tokenTypeUnaryOp, oper: tt.oper},
			}

			ops, stackSize := compileOps(rpn, map[string]int{}, make([]tOp, 0, 4))
			if len(ops) != 2 {
				t.Fatalf("len(ops) = %d, want 2", len(ops))
			}
			if stackSize != 1 {
				t.Fatalf("stackSize = %d, want 1", stackSize)
			}
			if ops[1].Cat != opCategoryUnary {
				t.Fatalf("ops[1].Cat = %v, want %v", ops[1].Cat, opCategoryUnary)
			}

			v := big.NewInt(tt.input)
			ops[1].Fn(v, v, nil)
			if v.Cmp(big.NewInt(tt.expected)) != 0 {
				t.Fatalf("result = %v, want %v", v, tt.expected)
			}
		})
	}
}

func TestLookupFn(t *testing.T) {
	cases := []struct {
		name    string
		wantNil bool
		wantCat tOpCat
	}{
		{name: "sqrt", wantCat: opCategoryUnary},
		{name: "abs", wantCat: opCategoryUnary},
		{name: "modInverse", wantCat: opCategoryBinary},
		{name: "modSqrt", wantCat: opCategoryBinary},
		{name: "quo", wantCat: opCategoryBinary},
		{name: "rem", wantCat: opCategoryBinary},
		{name: "missingFn", wantNil: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fn, cat := lookupFn(tt.name)
			if tt.wantNil {
				if fn != nil {
					t.Fatalf("lookupFn(%q) returned non-nil function", tt.name)
				}
				return
			}
			if fn == nil {
				t.Fatalf("lookupFn(%q) returned nil function", tt.name)
			}
			if cat != tt.wantCat {
				t.Fatalf("lookupFn(%q) category = %v, want %v", tt.name, cat, tt.wantCat)
			}
		})
	}
}

func TestTokenizeFunctionDetectionByParenthesis(t *testing.T) {
	tokens := tokenize("sqrt   (a) + sqrt", make([]tToken, 0, 8))

	var fnCount, paramSqrtCount int
	for _, tok := range tokens {
		if tok.val == "sqrt" && tok.tokenType == tokenTypeFn {
			fnCount++
		}
		if tok.val == "sqrt" && tok.tokenType == tokenTypeParam {
			paramSqrtCount++
		}
	}
	if fnCount != 1 || paramSqrtCount != 1 {
		t.Fatalf("expected one function sqrt and one param sqrt, got fn=%d param=%d", fnCount, paramSqrtCount)
	}
}

func TestCompilePanicsOnUnknownFunctionCall(t *testing.T) {
	panicVal := mustPanic(t, func() {
		_ = Compile("foo(a)", "a")
	})
	if !strings.Contains(panicVal.(string), "unknown function: foo") {
		t.Fatalf("unexpected panic: %v", panicVal)
	}
}

func TestCompilePanicsOnMismatchedRightParenthesis(t *testing.T) {
	_ = mustPanic(t, func() {
		_ = Compile("a + b)", "a", "b")
	})
}

func TestCompilePanicsOnInvalidCharacter(t *testing.T) {
	_ = mustPanic(t, func() {
		_ = Compile("a + , b", "a", "b")
	})
}

func TestEvalPanicsOnParamCountMismatch(t *testing.T) {
	p := Compile("a + b", "a", "b")
	_ = mustPanic(t, func() {
		_ = p.Exec(big.NewInt(1))
	})
}

func TestPrecedenceTable(t *testing.T) {
	tests := []struct {
		name     string
		oper     tOper
		expected int
	}{
		{name: "add", oper: operAdd, expected: 1},
		{name: "sub", oper: operSub, expected: 1},
		{name: "or", oper: operOr, expected: 1},
		{name: "xor", oper: operXor, expected: 1},
		{name: "mul", oper: operMul, expected: 2},
		{name: "div", oper: operDiv, expected: 2},
		{name: "mod", oper: operMod, expected: 2},
		{name: "and", oper: operAnd, expected: 2},
		{name: "and not", oper: operAndNot, expected: 2},
		{name: "none default", oper: operNone, expected: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := precedence(tt.oper); got != tt.expected {
				t.Fatalf("precedence(%v) = %d, want %d", tt.oper, got, tt.expected)
			}
		})
	}
}

func TestIsUnaryContext(t *testing.T) {
	tests := []struct {
		name      string
		tokensLen int
		prev      tTokenType
		expected  bool
	}{
		{name: "first token", tokensLen: 0, prev: tokenTypeNum, expected: true},
		{name: "after left paren", tokensLen: 1, prev: tokenTypeLParen, expected: true},
		{name: "after binary", tokensLen: 1, prev: tokenTypeBinaryOp, expected: true},
		{name: "after unary", tokensLen: 1, prev: tokenTypeUnaryOp, expected: true},
		{name: "after param", tokensLen: 1, prev: tokenTypeParam, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnaryContext(tt.tokensLen, tt.prev); got != tt.expected {
				t.Fatalf("isUnaryContext(%d, %v) = %v, want %v", tt.tokensLen, tt.prev, got, tt.expected)
			}
		})
	}
}

func TestCompileOpsPanicBranches(t *testing.T) {
	tests := []struct {
		name string
		rpn  []tToken
	}{
		{
			name: "invalid number token",
			rpn:  []tToken{{tokenType: tokenTypeNum, val: "12x"}},
		},
		{
			name: "unknown binary operator",
			rpn: []tToken{
				{tokenType: tokenTypeNum, val: "1"},
				{tokenType: tokenTypeNum, val: "2"},
				{tokenType: tokenTypeBinaryOp, oper: operNone},
			},
		},
		{
			name: "unknown unary operator",
			rpn: []tToken{
				{tokenType: tokenTypeNum, val: "1"},
				{tokenType: tokenTypeUnaryOp, oper: operNone},
			},
		},
		{
			name: "unhandled token type",
			rpn:  []tToken{{tokenType: tokenTypeLParen}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = mustPanic(t, func() {
				_, _ = compileOps(tt.rpn, map[string]int{}, make([]tOp, 0, 8))
			})
		})
	}
}

func TestToRPNBranches(t *testing.T) {
	t.Run("operator pop by precedence", func(t *testing.T) {
		in := []tToken{
			{tokenType: tokenTypeParam, val: "a"},
			{tokenType: tokenTypeBinaryOp, oper: operSub},
			{tokenType: tokenTypeParam, val: "b"},
			{tokenType: tokenTypeBinaryOp, oper: operSub},
			{tokenType: tokenTypeParam, val: "c"},
		}
		out := toRPN(in, make([]tToken, 0, 8))
		if len(out) != 5 {
			t.Fatalf("unexpected output len: %d", len(out))
		}
		if out[2].tokenType != tokenTypeBinaryOp || out[4].tokenType != tokenTypeBinaryOp {
			t.Fatalf("expected first operator to be popped before last operand, got %+v", out)
		}
	})

	t.Run("unary operators are emitted after operands", func(t *testing.T) {
		in := []tToken{
			{tokenType: tokenTypeUnaryOp, oper: operNeg},
			{tokenType: tokenTypeParam, val: "a"},
			{tokenType: tokenTypeBinaryOp, oper: operAdd},
			{tokenType: tokenTypeUnaryOp, oper: operNot},
			{tokenType: tokenTypeParam, val: "b"},
		}
		out := toRPN(in, make([]tToken, 0, 8))
		if len(out) != 5 {
			t.Fatalf("unexpected output len: %d", len(out))
		}
		if out[0].tokenType != tokenTypeParam || out[0].val != "a" {
			t.Fatalf("out[0] = %+v, want param a", out[0])
		}
		if out[1].tokenType != tokenTypeUnaryOp || out[1].oper != operNeg {
			t.Fatalf("out[1] = %+v, want unary neg", out[1])
		}
		if out[2].tokenType != tokenTypeParam || out[2].val != "b" {
			t.Fatalf("out[2] = %+v, want param b", out[2])
		}
		if out[3].tokenType != tokenTypeUnaryOp || out[3].oper != operNot {
			t.Fatalf("out[3] = %+v, want unary not", out[3])
		}
		if out[4].tokenType != tokenTypeBinaryOp || out[4].oper != operAdd {
			t.Fatalf("out[4] = %+v, want binary add", out[4])
		}
	})
}

func TestTokenizeInvalidUTF8Panics(t *testing.T) {
	invalid := string([]byte{0xff})
	tests := []struct {
		name string
		expr string
	}{
		{name: "invalid first rune", expr: invalid},
		{name: "invalid after and", expr: "&" + invalid},
		{name: "invalid in number tail", expr: "1" + invalid},
		{name: "invalid in ident tail", expr: "a" + invalid},
		{name: "invalid in function lookahead", expr: "foo " + invalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panicVal := mustPanic(t, func() {
				_ = tokenize(tt.expr, make([]tToken, 0, 8))
			})
			if !strings.Contains(fmt.Sprint(panicVal), "invalid utf-8 encoding") {
				t.Fatalf("unexpected panic: %v", panicVal)
			}
		})
	}
}

func TestCompilePanicsOnBinaryTilde(t *testing.T) {
	panicVal := mustPanic(t, func() {
		_ = Compile("a~b", "a", "b")
	})
	if !strings.Contains(fmt.Sprint(panicVal), "unknown operator code") {
		t.Fatalf("unexpected panic: %v", panicVal)
	}
}

func TestTokenizeUnaryPrefixOperators(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		wantOper   tOper
		secondType tTokenType
	}{
		{
			name:       "neg",
			expr:       "-a",
			wantOper:   operNeg,
			secondType: tokenTypeParam,
		},
		{
			name:       "not",
			expr:       "~a",
			wantOper:   operNot,
			secondType: tokenTypeParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.expr, make([]tToken, 0, 4))
			if len(tokens) != 2 {
				t.Fatalf("len(tokens)=%d, want 2", len(tokens))
			}
			if tokens[0].tokenType != tokenTypeUnaryOp || tokens[0].oper != tt.wantOper {
				t.Fatalf("first token = %+v, want unary %v", tokens[0], tt.wantOper)
			}
			if tokens[1].tokenType != tt.secondType || tokens[1].val != "a" {
				t.Fatalf("second token = %+v, want %v with val a", tokens[1], tt.secondType)
			}
		})
	}
}

func TestEvalReusesStackFromPool(t *testing.T) {
	p := Compile("(a + b) * c", "a", "b", "c")
	values := []*big.Int{big.NewInt(2), big.NewInt(3), big.NewInt(4)}

	got1 := p.Exec(values...)
	got2 := p.Exec(values...)
	want := big.NewInt(20)
	if got1.Cmp(want) != 0 || got2.Cmp(want) != 0 {
		t.Fatalf("Exec results = %v and %v, want %v", got1, got2, want)
	}
}

func mustPanic(t *testing.T, fn func()) any {
	t.Helper()
	var got any
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
				got = r
			}
		}()
		fn()
	}()
	if !didPanic {
		t.Fatal("expected panic, got nil")
	}
	return got
}

func BenchmarkCompile(b *testing.B) {
	expr := "(a + b) * (c - d) / (e + f)"
	vars := []string{"a", "b", "c", "d", "e", "f"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Compile(expr, vars...)
	}
}

func BenchmarkEval(b *testing.B) {
	expr := "(a + b) * (c - d) / (e + f)"
	vars := []string{"a", "b", "c", "d", "e", "f"}

	p := Compile(expr, vars...)

	values := []*big.Int{
		big.NewInt(100),
		big.NewInt(200),
		big.NewInt(300),
		big.NewInt(50),
		big.NewInt(10),
		big.NewInt(5),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Exec(values...)
	}
}

func BenchmarkEvalComplex(b *testing.B) {
	expr := "((a + b) * (c - d) / (e + f)) % g + (h * i) - (j / k)"
	vars := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	p := Compile(expr, vars...)

	values := []*big.Int{
		big.NewInt(100),
		big.NewInt(200),
		big.NewInt(300),
		big.NewInt(50),
		big.NewInt(10),
		big.NewInt(5),
		big.NewInt(7),
		big.NewInt(8),
		big.NewInt(9),
		big.NewInt(100),
		big.NewInt(25),
	}

	expected := big.NewInt(70)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := p.Exec(values...)
		if result.Cmp(expected) != 0 {
			b.Fatalf("mismatch: got %v, want %v", result, expected)
		}
	}
}

func BenchmarkEvalComplexRaw(b *testing.B) {
	values := []*big.Int{
		big.NewInt(100),
		big.NewInt(200),
		big.NewInt(300),
		big.NewInt(50),
		big.NewInt(10),
		big.NewInt(5),
		big.NewInt(7),
		big.NewInt(8),
		big.NewInt(9),
		big.NewInt(100),
		big.NewInt(25),
	}

	expected := big.NewInt(70)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		result := evalComplexRaw(values...)
		if result.Cmp(expected) != 0 {
			b.Fatalf("mismatch: got %v, want %v", result, expected)
		}
	}
}

func evalComplexRaw(values ...*big.Int) *big.Int {
	a := values[0]
	b := values[1]
	c := values[2]
	d := values[3]
	e := values[4]
	f := values[5]
	g := values[6]
	h := values[7]
	i := values[8]
	j := values[9]
	k := values[10]

	tmps := evalComplexRawPool.Get().(*[3]big.Int)
	defer evalComplexRawPool.Put(tmps)
	tmp1 := &tmps[0]
	tmp2 := &tmps[1]
	tmp3 := &tmps[2]

	tmp1.Add(a, b)
	tmp2.Sub(c, d)
	tmp1.Mul(tmp1, tmp2)
	tmp2.Add(e, f)
	tmp1.Div(tmp1, tmp2)
	tmp1.Mod(tmp1, g)
	tmp2.Mul(h, i)
	tmp1.Add(tmp1, tmp2)
	tmp3.Div(j, k)
	tmp1.Sub(tmp1, tmp3)

	return new(big.Int).Set(tmp1)
}
