package main

import (
	"text/template"
)

var lexTmpl = template.Must(template.New("yylex").Parse(`
{{- $yy := .Prefix -}}
// Generated from {{.Filename}}.  DO NOT EDIT.

{{printf "%s" .Prologue}}
type {{$yy}}Lex struct {
	Start   int32 // start condition
	Path    string
	Pos     int   // position of current token
	In      io.Reader
	buf     []byte
	linePos []int
	s, t    int // buf[s:t] == token to be flushed
	r, w    int // buf[r:w] == buffered text
	err     error
{{printf "%s" .Defscode}}
}

func (l *{{$yy}}Lex) Init(r io.Reader) *{{$yy}}Lex {
	l.Start = 0
	l.Pos = 0
	l.In = r
	l.buf = make([]byte, 4096)
	l.s, l.t, l.r, l.w = 0, 0, 0, 0
	l.err = nil
	return l
}

func (l *{{$yy}}Lex) ErrorAt(pos int, s string, v ...interface{}) {
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

func (l *{{$yy}}Lex) Error(s string) {
	l.ErrorAt(l.Pos, s)
}

func (l *{{$yy}}Lex) fill() {
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
			l.linePos = append(l.linePos, l.Pos + (i - l.s))
		}
	}
	l.w += n
	if err != nil {
		l.err = err
	}
}

func (l *{{$yy}}Lex) next() rune {
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

func ({{$yy}}lex *{{$yy}}Lex) Lex({{$yy}}lval *{{$yy}}SymType) int {
	const (
		{{- range $i, $v := .Starts }}
		{{.}}{{if not $i}} = iota{{end}} 
		{{- end}}
	)
	BEGIN := func(s int32) int32 {
		{{$yy}}lex.Start, s = s, {{$yy}}lex.Start
		return s
	}
	_ = BEGIN
	{{$yy}}less := func(n int) {
		n += {{$yy}}lex.s 
		{{$yy}}lex.t = n
		{{$yy}}lex.r = n
	}
	_ = {{$yy}}less
	{{$yy}}more := func() { {{$yy}}lex.t = {{$yy}}lex.s }
	_ = {{$yy}}more
{{printf "%s" .Rulescode}}
{{- range $i, $s := .States}}
{{$yy}}S{{$i}}:
{{- if eq $i 0}}
	{{$yy}}lex.Pos += {{$yy}}lex.t - {{$yy}}lex.s
	{{$yy}}lex.s = {{$yy}}lex.t
	{{$yy}}acc := -1
	{{$yy}}lex.t = {{$yy}}lex.r
	{{$yy}}c := {{$yy}}lex.Start
{{- else}}
{{- if ne .Tag -1}}
	{{$yy}}acc = {{.Tag}}
	{{$yy}}lex.t = {{$yy}}lex.r
{{- end}}
{{- if .Edges }}
	{{$yy}}c = {{$yy}}lex.next()
{{- end}}
{{- end}}
{{.DumpIfs $yy }}
	goto {{$yy}}fin
{{- end}}

{{$yy}}fin:
	{{$yy}}lex.r = {{$yy}}lex.t // put back read-ahead bytes
	{{$yy}}text := {{$yy}}lex.buf[{{$yy}}lex.s:{{$yy}}lex.r]
	if len({{$yy}}text) == 0 {
		if {{$yy}}lex.err != nil {
			return 0
		}
		panic("scanner is jammed")
	}
	switch {{$yy}}acc {
	{{- range $i, $r := .Rules}}
	{{- with $r.Action}}{{if ne . ""}}
	case {{$i}}:
		{{$r.Action}}
	{{- end}}{{end}}
	{{- end}}
	}
	goto {{$yy}}S0
}
`))
