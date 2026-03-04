package big

import (
	"math/big"
	"testing"
)

func TestExpressionExamples(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		vars     []string
		values   []*big.Int
		expected *big.Int
	}{
		{
			name:     "arithmetic",
			expr:     "(a + b) * (c - d) / 10 % m",
			vars:     []string{"a", "b", "c", "d", "m"},
			values:   []*big.Int{big.NewInt(11), big.NewInt(6), big.NewInt(19), big.NewInt(4), big.NewInt(9)},
			expected: big.NewInt(7),
		},
		{
			name:     "bitwise",
			expr:     "(a & b) | (c ^ d) &^ mask",
			vars:     []string{"a", "b", "c", "d", "mask"},
			values:   []*big.Int{big.NewInt(12), big.NewInt(10), big.NewInt(5), big.NewInt(3), big.NewInt(2)},
			expected: big.NewInt(12),
		},
		{
			name:     "shift",
			expr:     "(a << n) + (b >> n)",
			vars:     []string{"a", "b", "n"},
			values:   []*big.Int{big.NewInt(3), big.NewInt(16), big.NewInt(2)},
			expected: big.NewInt(16),
		},
		{
			name:     "unary",
			expr:     "-x + ~y",
			vars:     []string{"x", "y"},
			values:   []*big.Int{big.NewInt(5), big.NewInt(2)},
			expected: big.NewInt(-8),
		},
		{
			name:     "abs and sqrt",
			expr:     "abs(a) + sqrt(b)",
			vars:     []string{"a", "b"},
			values:   []*big.Int{big.NewInt(-12), big.NewInt(81)},
			expected: big.NewInt(21),
		},
		{
			name:     "modInverse rem quo",
			expr:     "modInverse(a, m) + rem(x, y) + quo(p, q)",
			vars:     []string{"a", "m", "x", "y", "p", "q"},
			values:   []*big.Int{big.NewInt(3), big.NewInt(11), big.NewInt(20), big.NewInt(6), big.NewInt(20), big.NewInt(6)},
			expected: big.NewInt(9),
		},
		{
			name:     "mixed",
			expr:     "abs(a-b) + ((c & d) ^ ~e) % m",
			vars:     []string{"a", "b", "c", "d", "e", "m"},
			values:   []*big.Int{big.NewInt(5), big.NewInt(12), big.NewInt(14), big.NewInt(11), big.NewInt(0), big.NewInt(7)},
			expected: big.NewInt(10),
		},
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
}
