package main

import (
	"fmt"
	"math/big"

	bigint "github.com/grudzinski/bigint"
)

var p = bigint.Compile(
	"((a + b) * (c - d) / (e + f)) % g + (h * i) - (j / k)",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
)

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

func evalWithBigint(a, b, c, d, e, f, g, h, i, j, k *big.Int) *big.Int {
	return p.Exec(a, b, c, d, e, f, g, h, i, j, k)
}

func main() {
	a := big.NewInt(100)
	b := big.NewInt(200)
	c := big.NewInt(300)
	d := big.NewInt(50)
	e := big.NewInt(10)
	f := big.NewInt(5)
	g := big.NewInt(7)
	h := big.NewInt(8)
	i := big.NewInt(9)
	j := big.NewInt(100)
	k := big.NewInt(25)

	raw := evalRaw(a, b, c, d, e, f, g, h, i, j, k)
	viaBigint := evalWithBigint(a, b, c, d, e, f, g, h, i, j, k)

	fmt.Printf("raw=%s bigint=%s\n", raw.String(), viaBigint.String())
	if raw.Cmp(viaBigint) != 0 {
		panic("mismatch between raw and bigint evaluation")
	}
}
