package main

import (
	"fmt"
	"math/big"
)

// Tokens
const (
	_ = iota + 2 // eof, error, unk
	NUM
	LPAR
	RPAR
	NL
	PLUS
	MINUS
	TIMES
	DIV
)

var yyName = []string{
	"$end",
	"error",
	"$unk",
	"NUM",
	"LPAR",
	"RPAR",
	"NL",
	"PLUS",
	"MINUS",
	"TIMES",
	"DIV",
}

const yyAccept = 1
const yyLast = 11

// Parse tables
var yyR1 = [...]int{
	0, 12, 12, 12, 11, 11, 11, 11, 11, 11,
	11, 11,
}

var yyR2 = [...]int{
	2, 0, 3, 3, 1, 3, 2, 2, 3, 3,
	3, 3,
}

var yyReduce = [...]int{
	1, 0, 0, 4, 0, 0, 0, 0, 3, 0,
	6, 7, 2, 0, 0, 0, 0, 5, 8, 9,
	10, 11,
}

var yyGoto = [...]int{
	7, 1,
}

var yyAction = [...]int{
	9, 10, 11, 17, 0, 13, 14, 15, 16, 18,
	19, 20, 21, 2, 0, 3, 4, 15, 16, 5,
	6, 12, 13, 14, 15, 16, 3, 4, 8, 0,
	5, 6,
}

var yyCheck = [...]int{
	4, 5, 6, 5, -1, 7, 8, 9, 10, 13,
	14, 15, 16, 1, -1, 3, 4, 9, 10, 7,
	8, 6, 7, 8, 9, 10, 3, 4, 6, -1,
	7, 8,
}

var yyPact = [...]int{
	32, 12, 22, 32, 23, 23, 23, 15, 32, -2,
	8, 8, 32, 23, 23, 23, 23, 32, 8, 8,
	32, 32,
}

var yyPgoto = [...]int{
	-4, 32,
}

type yySymType struct {
	yys int // state

	num *big.Rat

}

type yyLexer interface {
	Lex(*yySymType) int
	Error(string)
}

var yyDebug = 0 // debug info from parser

// yyParse read tokens from yylex and parses input.
// Returns result on success, or nil on failure.
func yyParse(yylex yyLexer) *yySymType {
	var (
		yyn, yyt int
		yystate  = 0
		yyerror  = 0
		yymajor  = -1
		yystack  []yySymType
		yyD      []yySymType // rhs of reduction
		yylval   yySymType   // lexcial value from lexer
		yyval    yySymType   // value to be pushed onto stack
	)
	goto yyaction
yystack:
	yyval.yys = yystate
	yystack = append(yystack, yyval)
	yystate = yyn
	if yyDebug >= 2 {
		println("\tGOTO state", yyn)
	}
yyaction:
	// look up shift or reduce
	yyn = int(yyPact[yystate])
	if yyn == len(yyAction) && yystate != yyAccept { // simple state
		goto yydefault
	}
	if yymajor < 0 {
		yymajor = yylex.Lex(&yylval)
		if yyDebug >= 1 {
			println("In state", yystate)
		}
		if yyDebug >= 2 {
			println("\tInput token", yyName[yymajor])
		}
	}
	yyn += yymajor
	if 0 <= yyn && yyn < len(yyAction) && int(yyCheck[yyn]) == yymajor {
		yyn = int(yyAction[yyn])
		if yyn <= 0 {
			yyn = -yyn
			goto yyreduce
		}
		if yyDebug >= 1 {
			println("\tSHIFT token", yyName[yymajor])
		}
		if yyerror > 0 {
			yyerror--
		}
		yymajor = -1
		yyval = yylval
		goto yystack
	}
yydefault:
	yyn = int(yyReduce[yystate])
yyreduce:
	if yyn == 0 {
		if yymajor == 0 && yystate == yyAccept {
			if yyDebug >= 1 {
				println("\tACCEPT!")
			}
			return &yystack[0]
		}
		switch yyerror {
		case 0: // new error
			if yyDebug >= 1 {
				println("\tERROR!")
			}
			msg := "unexpected " + yyName[yymajor]
			var expect []int
			if yyReduce[yystate] == 0 {
				yyn = yyPact[yystate] + 3
				for i := 3; i < yyLast; i++ {
					if 0 <= yyn && yyn < len(yyAction) && yyCheck[yyn] == i && yyAction[yyn] != 0 {
						expect = append(expect, i)
						if len(expect) > 4 {
							break
						}
					}
					yyn++
				}
			}
			if n := len(expect); 0 < n && n <= 4 {
				for i, tok := range expect {
					switch i {
					case 0:
						msg += ", expecting "
					case n-1:
						msg += " or "
					default:
						msg += ", "
					}
					msg += yyName[tok]
				}
			}
			yylex.Error(msg)
			fallthrough
		case 1, 2: // partially recovered error
			for { // pop states until error can be shifted
				yyn = int(yyPact[yystate]) + 1
				if 0 <= yyn && yyn < len(yyAction) && yyCheck[yyn] == 1 {
					yyn = yyAction[yyn]
					if yyn > 0 {
						break
					}
				}
				if len(yystack) == 0 {
					return nil
				}
				if yyDebug >= 2 {
					println("\tPopping state", yystate)
				}
				yystate = yystack[len(yystack)-1].yys
				yystack = yystack[:len(yystack)-1]
			}
			yyerror = 3
			if yyDebug >= 1 {
				println("\tSHIFT token error")
			}
			goto yystack
		default: // still waiting for valid tokens
			if yymajor == 0 { // no more tokens
				return nil
			}
			if yyDebug >= 1 {
				println("\tDISCARD token", yyName[yymajor])
			}
			yymajor = -1
			goto yyaction
		}
	}
	if yyDebug >= 1 {
		println("\tREDUCE rule", yyn)
	}
	yyt = len(yystack) - int(yyR2[yyn])
	yyD = yystack[yyt:]
	if len(yyD) > 0 { // pop items and restore state
		yyval = yyD[0]
		yystate = yyval.yys
		yystack = yystack[:yyt]
	}
	switch yyn { // Semantic actions

	case 2:

		var v interface{}
		if yyD[1].num.IsInt() {
			v = yyD[1].num.Num()
		} else {
			v = yyD[1].num
		}
		fmt.Println(v)
	
	case 3:
 yyerror = 0 
	case 5:
 yyval.num = yyD[1].num 
	case 6:
 yyval.num = yyD[1].num 
	case 7:
 yyval.num = yyD[1].num.Neg(yyD[1].num) 
	case 8:
 yyval.num = yyD[0].num.Add(yyD[0].num, yyD[2].num) 
	case 9:
 yyval.num = yyD[0].num.Sub(yyD[0].num, yyD[2].num) 
	case 10:
 yyval.num = yyD[0].num.Mul(yyD[0].num, yyD[2].num) 
	case 11:
 yyval.num = yyD[0].num.Quo(yyD[0].num, yyD[2].num) 
	}
	// look up goto
	yyt = int(yyR1[yyn]) - yyLast
	yyn = int(yyPgoto[yyt]) + yystate
	if 0 <= yyn && yyn < len(yyAction) &&
		int(yyCheck[yyn]) == yystate {
		yyn = int(yyAction[yyn])
	} else {
		yyn = int(yyGoto[yyt])
	}
	goto yystack
}
