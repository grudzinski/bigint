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
