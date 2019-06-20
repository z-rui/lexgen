package main

import (
	"math/big"
)

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
