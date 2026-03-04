// Package big simplifies working with math/big.Int by evaluating string expressions.
//
// It lets you declare expressions once and evaluate them repeatedly with different
// input values.
// It also improves readability and maintainability versus equivalent
// hand-written math/big mutation chains.
// In typical repeated-use workloads, this usually improves runtime and memory
// usage compared to equivalent hand-written math/big mutation chains.
//
// # How to Use
//
// Write your expression as a string, for example "(a + b) * c", then call [Compile]
// (or [Prog.Init]) with parameter names in the same order you will pass values.
// Execute the compiled program with [Prog.Exec], passing *big.Int values in that
// exact order.
//
// Compile once and reuse the same [Prog] for many [Prog.Exec] calls.
//
// # Syntax
//
// The language supports binary operators +, -, *, /, %, <<, >>, &, |, ^, and
// &^, unary operators - (negation) and ~ (bitwise NOT), and parentheses.
// Operator precedence follows Go for supported operators: multiplicative
// (*, /, %, <<, >>, &, &^) binds tighter than additive (+, -, |, ^).
//
// # Functions
//
// Supported functions are sqrt(x), abs(x), modInverse(x, y), modSqrt(x, y),
// quo(x, y), and rem(x, y).
//
// # Expression Examples
//
//	(a + b) * (c - d) / 10 % m
//	(a & b) | (c ^ d) &^ mask
//	(a << n) + (b >> n)
//	-x + ~y
//	abs(a) + sqrt(b)
//	modInverse(a, m) + rem(x, y) + quo(p, q)
//	abs(a-b) + ((c & d) ^ ~e) % m
//
// # Quick Start
//
//	import (
//		"fmt"
//		"math/big"
//
//		bigint "github.com/grudzinski/bigint"
//	)
//
//	func main() {
//		p := bigint.Compile("(a + b) * c", "a", "b", "c")
//		result := p.Exec(big.NewInt(2), big.NewInt(3), big.NewInt(4))
//		fmt.Println(result) // 20
//	}
//
// # Comparison
//
//	expr := "((a + b) * (c - d) / (e + f)) % g + (h * i) - (j / k)"
//
//	// Without bigint (raw math/big.Int chains):
//	func evalRaw(a, b, c, d, e, f, g, h, i, j, k *big.Int) *big.Int {
//		tmp1 := new(big.Int).Add(a, b)
//		tmp2 := new(big.Int).Sub(c, d)
//		tmp1.Mul(tmp1, tmp2)
//		tmp2.Add(e, f)
//		tmp1.Div(tmp1, tmp2)
//		tmp1.Mod(tmp1, g)
//		tmp2.Mul(h, i)
//		tmp1.Add(tmp1, tmp2)
//		tmp3 := new(big.Int).Div(j, k)
//		tmp1.Sub(tmp1, tmp3)
//		return tmp1
//	}
//
//	// With bigint:
//	var p = bigint.Compile(
//		expr,
//		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
//	)
//
//	func evalWithBigint(a, b, c, d, e, f, g, h, i, j, k *big.Int) *big.Int {
//		return p.Exec(a, b, c, d, e, f, g, h, i, j, k)
//	}
package big
