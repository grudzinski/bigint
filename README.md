# bigint

`bigint` simplifies working with `math/big.Int` by letting you write expressions instead of manual mutation chains.

It is designed for repeated evaluation with different input values.
It improves readability and maintainability versus equivalent hand-written
`math/big` mutation chains.

## Comparison: `math/big` vs `bigint`

Example expression:

`((a + b) * (c - d) / (e + f)) % g + (h * i) - (j / k)`

Without `bigint` (direct `math/big`):

```go
func evalRaw(a, b, c, d, e, f, g, h, i, j, k *big.Int) *big.Int {
	tmp1 := new(big.Int).Add(a, b)
	tmp2 := new(big.Int).Sub(c, d)
	tmp1.Mul(tmp1, tmp2)

	tmp2.Add(e, f)
	tmp1.Div(tmp1, tmp2)
	tmp1.Mod(tmp1, g)

	tmp2.Mul(h, i)
	tmp1.Add(tmp1, tmp2)

	tmp3 := new(big.Int).Div(j, k)
	tmp1.Sub(tmp1, tmp3)
	return tmp1
}
```

With `bigint`:

```go
var p = bigint.Compile(
	"((a + b) * (c - d) / (e + f)) % g + (h * i) - (j / k)",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
)

func evalWithBigint(a, b, c, d, e, f, g, h, i, j, k *big.Int) *big.Int {
	return p.Exec(a, b, c, d, e, f, g, h, i, j, k)
}
```

## Why

Using `math/big` directly for non-trivial expressions is verbose:

- many temporary values
- long mutation chains (`Add`, `Mul`, `Div`, ...)
- hard-to-read business expressions

`bigint` lets you declare an expression once and evaluate it many times:

- expression stays readable (`"(a + b) * c"`)
- readability and maintainability are higher than equivalent manual `math/big` code
- one compile step, fast repeated `Exec`
- still uses native `*big.Int` values
- in typical repeated-use workloads, `Exec` is faster and allocates less memory
  than equivalent hand-written `math/big` mutation chains

## Install

```bash
go get github.com/grudzinski/bigint
```

## How to Use

Write your expression as a string, for example "(a + b) * c", then call `Compile`
(or `Prog.Init`) with parameter names in the same order you will pass values.
Execute the compiled program with `Prog.Exec`, passing `*big.Int` values in that
exact order.

Compile once and reuse the same `Prog` for many `Prog.Exec` calls.

## Quick Start

```go
package main

import (
	"fmt"
	"math/big"

	bigint "github.com/grudzinski/bigint"
)

func main() {
	p := bigint.Compile("(a + b) * c", "a", "b", "c")

	result := p.Exec(
		big.NewInt(2),
		big.NewInt(3),
		big.NewInt(4),
	)

	fmt.Println(result) // 20
}
```

## Supported Syntax

The language supports binary operators `+`, `-`, `*`, `/`, `%`, `&`, `|`, `^`,
and `&^`, unary operators `-` (negation) and `~` (bitwise NOT), and parentheses.
Operator precedence follows Go for supported operators: multiplicative
(`*`, `/`, `%`, `&`, `&^`) binds tighter than additive (`+`, `-`, `|`, `^`).

Supported functions are `sqrt(x)`, `abs(x)`, `modInverse(x, y)`,
`modSqrt(x, y)`, `quo(x, y)`, and `rem(x, y)`.

### Expression examples

- `(a + b) * (c - d) / 10 % m`
- `(a & b) | (c ^ d) &^ mask`
- `-x + ~y`
- `abs(a) + sqrt(b)`
- `modInverse(a, m) + rem(x, y) + quo(p, q)`
- `abs(a-b) + ((c & d) ^ ~e) % m`

## API

- `Compile(expr string, paramNames ...string) *Prog`
- `(*Prog).Exec(paramVals ...*big.Int) *big.Int`
- `Exec` is safe for concurrent use by multiple goroutines after compilation.

## Behavior and Errors

The package uses `panic` for invalid expressions and runtime misuse, including:

- unknown parameter/function
- mismatched parentheses
- invalid token/character
- wrong number of `Exec` parameters

## License

MIT. See [LICENSE](LICENSE).
