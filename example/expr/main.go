package main

//go:generate lexgen expr.l
//go:generate lrgen expr.y

import (
	"os"
)

func main() {
	l := new(yyLex)
	l.Init(os.Stdin)
	l.Path = "<stdin>"
	yyParse(l)
}
