// Generated from expr.l.  DO NOT EDIT.

package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"unicode/utf8"
)

import (
	"math/big"
)

type yyLex struct {
	Start   int32 // start condition
	Path    string
	Pos     int // position of current token
	In      io.Reader
	buf     []byte
	linePos []int
	s, t    int // buf[s:t] == token to be flushed
	r, w    int // buf[r:w] == buffered text
	err     error
}

func (l *yyLex) Init(r io.Reader) *yyLex {
	l.Start = 0
	l.Pos = 0
	l.In = r
	l.buf = make([]byte, 4096)
	l.s, l.t, l.r, l.w = 0, 0, 0, 0
	l.err = nil
	return l
}

func (l *yyLex) ErrorAt(pos int, s string, v ...interface{}) {
	if len(v) > 0 {
		s = fmt.Sprintf(s, v...)
	}
	lin := sort.SearchInts(l.linePos, pos)
	col := pos
	if lin > 0 {
		col -= l.linePos[lin-1] + 1
	}
	fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", l.Path, lin+1, col+1, s)
}

func (l *yyLex) Error(s string) {
	l.ErrorAt(l.Pos, s)
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
	n, err := l.In.Read(l.buf[l.w:])
	// update newline positions
	for i := l.w; i < l.w+n; i++ {
		if l.buf[i] == '\n' {
			l.linePos = append(l.linePos, l.Pos+(i-l.s))
		}
	}
	l.w += n
	if err != nil {
		l.err = err
	}
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

func (yy *yyLex) Lex(yylval *yySymType) int {
	const (
		INITIAL = iota
	)
	BEGIN := func(s int32) int32 {
		yy.Start, s = s, yy.Start
		return s
	}
	_ = BEGIN
	yyless := func(n int) {
		n += yy.s
		yy.t = n
		yy.r = n
	}
	_ = yyless
	yymore := func() { yy.t = yy.s }
	_ = yymore

yyS0:
	yy.Pos += yy.t - yy.s
	yy.s = yy.t
	yyacc := -1
	yy.t = yy.r
	yyc := yy.Start
	if '\x00' <= yyc && yyc <= '\x00' {
		goto yyS1
	}

	goto yyfin
yyS1:
	yyc = yy.next()
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
	yy.t = yy.r

	goto yyfin
yyS3:
	yyacc = 7
	yy.t = yy.r

	goto yyfin
yyS4:
	yyacc = 6
	yy.t = yy.r

	goto yyfin
yyS5:
	yyacc = 4
	yy.t = yy.r

	goto yyfin
yyS6:
	yyacc = 5
	yy.t = yy.r

	goto yyfin
yyS7:
	yyacc = 2
	yy.t = yy.r

	goto yyfin
yyS8:
	yyacc = 0
	yy.t = yy.r

	goto yyfin
yyS9:
	yyacc = 1
	yy.t = yy.r

	goto yyfin
yyS10:
	yyacc = 9
	yy.t = yy.r
	yyc = yy.next()
	if '0' <= yyc && yyc <= '9' {
		goto yyS13
	}

	goto yyfin
yyS11:
	yyacc = 3
	yy.t = yy.r

	goto yyfin
yyS12:
	yyacc = 8
	yy.t = yy.r
	yyc = yy.next()
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
	yy.t = yy.r
	yyc = yy.next()
	if '0' <= yyc && yyc <= '9' {
		goto yyS13
	}

	goto yyfin

yyfin:
	yy.r = yy.t // put back read-ahead bytes
	yytext := yy.buf[yy.s:yy.r]
	if len(yytext) == 0 {
		if yy.err != nil {
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
