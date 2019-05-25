package main

import (
	"fmt"
	"math/big"
	"os"
)

func (l *yyLex) Error(s string) {
	fmt.Fprintf(os.Stderr, "%s\n", s)
}

%%

%%

"+"	return PLUS
"-"	return MINUS
[*×]	return TIMES
[/÷]	return DIV
"("	return LPAR
")"	return RPAR
\n	return NL
[ \t\r\f\v]
[0-9]+(|\.[0-9]*)|\.[0-9]+	{
	yylval.num, _ = new(big.Rat).SetString(string(yytext))
	return NUM
}
.	return 2
