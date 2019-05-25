package main

//go:generate lexgen expr.l
//go:generate lrgen expr.y

import (
	"os"
)

func main() {
	yyParse(yyNewLexer(os.Stdin))
}
