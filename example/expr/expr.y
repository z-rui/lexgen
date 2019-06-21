package main

import (
	"fmt"
	"math/big"
)

%%

%union {
	num *big.Rat
}

%token <num> NUM
%token LPAR RPAR NL

%left PLUS MINUS
%left TIMES DIV

%type <num> expr

%%

input:
	/* epsilon */
|	input expr NL
	{
		var v interface{}
		if $2.IsInt() {
			v = $2.Num()
		} else {
			v = $2
		}
		fmt.Println(v)
	}
|	input error NL { yyerror = 0 }
;

expr:
	NUM
|	LPAR expr RPAR	{ $$ = $2 }
|	PLUS expr	{ $$ = $2 }
|	MINUS expr	{ $$ = $2.Neg($2) }
|	expr PLUS expr	{ $$ = $1.Add($1, $3) }
|	expr MINUS expr	{ $$ = $1.Sub($1, $3) }
|	expr TIMES expr	{ $$ = $1.Mul($1, $3) }
|	expr DIV expr
	{
		if $3.Sign() == 0 {
			yy.ErrorAt(@2, "division by zero")
		} else {
			$$ = $1.Quo($1, $3)
		}
	}
;
