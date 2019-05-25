// Generated from expr.l.  DO NOT EDIT.

package main

import (
	"io"
	"unicode/utf8"
)

import (
	"fmt"
	"math/big"
	"os"
)

func (l *yyLex) Error(s string) {
	fmt.Fprintf(os.Stderr, "%s\n", s)
}

type yyLex struct {
	Start int32 // start condition
	rd    io.Reader
	buf   []byte
	pos   int
	s, t  int // buf[s:t] == token to be flushed
	r, w  int // buf[r:w] == buffered text
	err   error
}

func yyNewLexer(r io.Reader) *yyLex {
	return &yyLex{
		rd: r,
	}
}

func (l *yyLex) fill() {
	if n := len(l.buf); l.w == n {
		if l.s+l.s <= len(l.buf) {
			// less than half usable, better extend buffer
			if n == 0 {
				n = 4096
			} else {
				n *= 2
			}
			buf := make([]byte, n)
			copy(buf, l.buf[l.s:])
			l.buf = buf
		} else {
			// shift content
			copy(l.buf, l.buf[l.s:])
		}
		l.r -= l.s
		l.w -= l.s
		l.t -= l.s
		l.s = 0
	}
	n, err := l.rd.Read(l.buf[l.w:])
	l.w += n
	if err != nil {
		l.err = err
	}
}

func (l *yyLex) shift() {
}

func (l *yyLex) next() rune {
	for l.r+utf8.UTFMax > l.w && !utf8.FullRune(l.buf[l.r:l.w]) && l.err == nil {
		l.fill()
	}
	if l.r == l.w { // nothing is available
		return -1
	}
	c, n := rune(l.buf[l.r]), 1
	if c >= utf8.RuneSelf {
		c, n = utf8.DecodeRune(l.buf[l.r:l.w])
	}
	l.r += n
	return c
}

func (yylex *yyLex) Lex(yylval *yySymType) int {
	const (
		INITIAL = iota
	)
	BEGIN := func(s int32) int32 {
		yylex.Start, s = s, yylex.Start
		return s
	}
	BEGIN(INITIAL)
	yyless := func(n int) {
		n += yylex.s
		yylex.t = n
		yylex.r = n
	}
	_ = yyless
	yymore := func() { yylex.t = yylex.s }
	_ = yymore

yyS0:
	yylex.pos += yylex.t - yylex.s
	yylex.s = yylex.t
	yyacc := -1
	yylex.t = yylex.r
	yyc := yylex.Start
	if '\x00' <= yyc && yyc <= '\x00' {
		goto yyS1
	}

	goto yyfin
yyS1:
	yyc = yylex.next()
	if yyc < '+' {
		if yyc < ' ' {
			if yyc < '\n' {
				if yyc < '\t' {
					if '\x00' <= yyc {
						goto yyS2
					}
				} else {
					goto yyS3
				}
			} else if yyc < '\v' {
				goto yyS4
			} else if yyc < '\x0e' {
				goto yyS3
			} else {
				goto yyS2
			}
		} else if yyc < '(' {
			if yyc < '!' {
				goto yyS3
			} else {
				goto yyS2
			}
		} else if yyc < ')' {
			goto yyS5
		} else if yyc < '*' {
			goto yyS6
		} else {
			goto yyS7
		}
	} else if yyc < '0' {
		if yyc < '-' {
			if yyc < ',' {
				goto yyS8
			} else {
				goto yyS2
			}
		} else if yyc < '.' {
			goto yyS9
		} else if yyc < '/' {
			goto yyS10
		} else {
			goto yyS11
		}
	} else if yyc < 'Ø' {
		if yyc < ':' {
			goto yyS12
		} else if yyc < '×' {
			goto yyS2
		} else {
			goto yyS7
		}
	} else if yyc < '÷' {
		goto yyS2
	} else if yyc < 'ø' {
		goto yyS11
	} else if yyc <= '\U0010ffff' {
		goto yyS2
	}

	goto yyfin
yyS2:
	yyacc = 9
	yylex.t = yylex.r

	goto yyfin
yyS3:
	yyacc = 7
	yylex.t = yylex.r

	goto yyfin
yyS4:
	yyacc = 6
	yylex.t = yylex.r

	goto yyfin
yyS5:
	yyacc = 4
	yylex.t = yylex.r

	goto yyfin
yyS6:
	yyacc = 5
	yylex.t = yylex.r

	goto yyfin
yyS7:
	yyacc = 2
	yylex.t = yylex.r

	goto yyfin
yyS8:
	yyacc = 0
	yylex.t = yylex.r

	goto yyfin
yyS9:
	yyacc = 1
	yylex.t = yylex.r

	goto yyfin
yyS10:
	yyacc = 9
	yylex.t = yylex.r
	yyc = yylex.next()
	if '0' <= yyc && yyc <= '9' {
		goto yyS13
	}

	goto yyfin
yyS11:
	yyacc = 3
	yylex.t = yylex.r

	goto yyfin
yyS12:
	yyacc = 8
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '0' {
		if '.' <= yyc && yyc <= '.' {
			goto yyS13
		}
	} else if yyc <= '9' {
		goto yyS12
	}

	goto yyfin
yyS13:
	yyacc = 8
	yylex.t = yylex.r
	yyc = yylex.next()
	if '0' <= yyc && yyc <= '9' {
		goto yyS13
	}

	goto yyfin

yyfin:
	yylex.r = yylex.t // put back read-ahead bytes
	yytext := yylex.buf[yylex.s:yylex.r]
	if len(yytext) == 0 {
		if yylex.err != nil {
			return 0
		}
		panic("scanner is jammed")
	}
	switch yyacc {
	case 0:
		return PLUS
	case 1:
		return MINUS
	case 2:
		return TIMES
	case 3:
		return DIV
	case 4:
		return LPAR
	case 5:
		return RPAR
	case 6:
		return NL
	case 8:
		{
			yylval.num, _ = new(big.Rat).SetString(string(yytext))
			return NUM
		}
	case 9:
		return 2
	}
	goto yyS0
}
