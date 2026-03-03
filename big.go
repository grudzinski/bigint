package big

// Compile creates a new [Prog], calls [Prog.Init] with expr and paramNames, and
// returns the compiled program. The returned [Prog] can be executed repeatedly
// with different *big.Int values. Compile panics under the same conditions as
// [Prog.Init].
func Compile(expr string, paramNames ...string) *Prog {
	p := new(Prog)
	p.Init(expr, paramNames...)
	return p
}
