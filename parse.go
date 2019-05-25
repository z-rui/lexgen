package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
)

const maxErrors = 10

type rule struct {
	frag
	Start  rnSet
	Action string
}

type parseResult struct {
	mach
	Prefix    string
	Prologue  []byte
	Rules     []rule
	Defscode  []byte
	Rulescode []byte
	Starts    []string
}

type parser struct {
	rd *bufio.Reader
	wr *bufio.Writer

	// parser state
	Filename string
	line     int
	errors   int
	allowWS  bool

	defs   map[string]frag
	starts map[string]int32
	shared []int32

	parseResult
}

func (p *parser) error(s string) {
	fmt.Fprintf(os.Stderr, "%s:%d: %s\n", p.Filename, p.line, s)
	p.errors++
	if p.errors > maxErrors {
		fmt.Fprintf(os.Stderr, "too many errors\n")
		os.Exit(1)
	}
}

func (p *parser) errorf(s string, v ...interface{}) {
	p.error(fmt.Sprintf(s, v...))
}

func (p *parser) Parse(r io.Reader) {
	p.rd = bufio.NewReader(r)
	p.defs = map[string]frag{}
	p.starts = map[string]int32{"INITIAL": 0}
	p.shared = []int32{0}

	p.readPrologue()
	p.readDefs()
	p.readRules()

	p.Starts = make([]string, len(p.starts))
	for k, v := range p.starts {
		p.Starts[v] = k
	}
	p.makeMach()
}

func (p *parser) makeMach() {
	start := &fragState{}
	acc := map[*fragState]int{} // frag state -> rule id
	for i, r := range p.Rules {
		s := &fragState{Input: r.Start, Succ: []*fragState{r.frag.Start}}
		start.Succ = append(start.Succ, s)
		acc[r.Accept] = i
	}
	p.mach = newMach(frag{start, nil}, acc)
	p.Prefix = "yy"
}

func (p *parser) skipWS() {
	for {
		switch c, _ := p.rd.ReadByte(); c {
		case ' ', '\t':
		default:
			p.rd.UnreadByte()
			return
		}
	}
}

func (p *parser) scanIdent() string {
	var buf []byte
	c, _ := p.rd.ReadByte()
	switch {
	case 'A' <= c && c <= 'Z', 'a' <= c && c <= 'z', c == '_':
		buf = append(buf, c)
	default:
		goto out
	}
	for {
		c, _ := p.rd.ReadByte()
		switch {
		case '0' <= c && c <= '9', 'A' <= c && c <= 'Z', 'a' <= c && c <= 'z', c == '_':
			buf = append(buf, c)
		default:
			goto out
		}
	}
out:
	p.rd.UnreadByte()
	return string(buf)
}

func (p *parser) scanPaired(l, r byte) string {
	buf := []byte{}
	level := 0
L:
	for {
		c, err := p.rd.ReadByte()
		if err != nil {
			break
		}
		buf = append(buf, c)
		switch c {
		case r:
			if level--; level == 0 {
				break L
			}
		case l:
			level++
		}
	}
	return string(buf)
}

func (p *parser) readPrologue() {
	for {
		line, err := p.rd.ReadBytes('\n')
		p.line++
		if err != nil {
			break
		}
		if len(line) == 3 && string(line) == "%%\n" {
			break
		}
		p.Prologue = append(p.Prologue, line...)
	}
}

func (p *parser) readDefs() {
	for {
		c, err := p.rd.ReadByte()
		if err != nil {
			break
		}
		switch c {
		case '\n':
			p.line++
		case '%':
			s, _ := p.rd.Peek(2)
			if len(s) == 2 && s[0] == '%' && s[1] == '\n' {
				p.rd.Discard(2)
				p.line++
				return
			}
			p.readDirective()
		case ' ', '\t':
			line, _ := p.rd.ReadBytes('\n')
			p.line++
			p.Defscode = append(p.Defscode, line...)
		default:
			p.rd.UnreadByte()
			p.readDef()
		}
	}
}

func (p *parser) readDirective() {
	directive := p.scanIdent()
	var shared bool
	switch directive {
	case "s", "S":
		shared = true
	case "x", "X":
		shared = false
	default:
		p.error("bad directive")
		goto skipLine
	}
	for {
		switch c, _ := p.rd.ReadByte(); c {
		case ' ', '\t':
			p.skipWS()
		case '\n':
			p.line++
			return
		default:
			p.rd.UnreadByte()
		}
		name := p.scanIdent()
		if name == "" {
			p.error("expecting identifier")
			goto skipLine
		}
		if _, ok := p.starts[name]; ok {
			p.errorf("start condition %q redefined", name)
			continue
		}
		sc := int32(len(p.starts))
		p.starts[name] = sc
		if shared {
			p.shared = append(p.shared, sc)
		}
	}
skipLine:
	p.rd.ReadBytes('\n') // skip to LF
	p.line++
}

func (p *parser) readDef() {
	name := p.scanIdent()
	if name == "" {
		p.error("expecting identifier")
		goto skipLine
	}
	switch c, _ := p.rd.ReadByte(); c {
	case ' ', '\t':
		p.skipWS()
	case '\n':
		p.rd.UnreadByte()
	default:
		p.error("expecting whitespace")
		goto skipLine
	}
	p.defs[name] = p.parseRE()
	if c, _ := p.rd.ReadByte(); c != '\n' {
		p.error("expecting newline")
		goto skipLine
	}
	p.line++
	return
skipLine:
	p.rd.ReadBytes('\n') // skip to LF
	p.line++
}

func (p *parser) readRules() {
	shared := make(rnSet, len(p.shared))
	for i, sc := range p.shared {
		shared[i] = rn{sc, sc}
	}
	shared = shared.Canon(false)
	for {
		c, err := p.rd.ReadByte()
		if err != nil {
			break
		}
		switch c {
		case '\n':
			p.line++
		case ' ', '\t':
			line, _ := p.rd.ReadBytes('\n')
			p.line++
			p.Rulescode = append(p.Rulescode, line...)
		case '<':
			s, _ := p.rd.Peek(2)
			var start rnSet
			if len(s) == 2 && s[0] == '*' && s[1] == '>' {
				p.rd.Discard(2)
				start = rnSet{{0, int32(len(p.starts) - 1)}}
			} else {
				c := byte(',')
				for c == ',' {
					name := p.scanIdent()
					c, _ = p.rd.ReadByte()
					if name == "" {
						p.errorf("expecting identifier")
						goto skipLine
					}
					if sc, ok := p.starts[name]; ok {
						start = append(start, rn{sc, sc})
					} else {
						p.errorf("%q undefined", name)
					}
				}
				if c != '>' {
					p.errorf("expecting '>'")
					goto skipLine
				}
				start = start.Canon(false)
			}
			p.readRule(start)
		default:
			p.rd.UnreadByte()
			p.readRule(shared)
		}
	}
	return
skipLine:
	p.rd.ReadBytes('\n') // skip to LF
	p.line++
}

func (p *parser) readRule(start rnSet) {
	var rule rule
	rule.frag = p.parseRE()
	rule.Start = start

	switch c, _ := p.rd.ReadByte(); c {
	case ' ', '\t':
		p.skipWS()
	case '\n':
		p.rd.UnreadByte()
	default:
		p.error("expecting whitespace")
		goto skipLine
	}
	if s, _ := p.rd.Peek(1); len(s) == 1 && s[0] == '{' {
		rule.Action = p.scanPaired('{', '}')
		if c, _ := p.rd.ReadByte(); c != '\n' {
			p.error("expecting newline")
			goto skipLine
		}
	} else {
		s, _ = p.rd.ReadBytes('\n')
		if len(s) > 0 {
			s = s[:len(s)-1]
		}
		rule.Action = string(s)
	}
	p.line++
	p.Rules = append(p.Rules, rule)
	return
skipLine:
	p.rd.ReadBytes('\n') // skip to LF
	p.line++
}

func (p *parser) parseRE() frag {
	// re : concats { "|" concats }
	frags := []frag{p.parseConcats()}
	for {
		if s, _ := p.rd.Peek(1); len(s) == 1 && s[0] == '|' {
			p.rd.Discard(1)
			frags = append(frags, p.parseConcats())
		} else {
			break
		}
	}
	return Alter(frags)
}

func (p *parser) parseConcats() frag {
	// concats : { primary }
	frags := []frag{}
L:
	for {
		s, _ := p.rd.Peek(1)
		if len(s) < 1 {
			break
		}
		switch s[0] {
		case '\n', '|', ')':
			break L
		case ' ', '\t':
			if !p.allowWS {
				break L
			}
		}
		frags = append(frags, p.parsePrimary())
	}
	return Concat(frags)
}

func (p *parser) parsePrimary() frag {
	/* primary
	: CHARSET
	| NAME
	| LITERAL
	| "(" re ")"
	| primary "*"
	| primary "+"
	*/
	var frag frag
	c, _ := p.rd.ReadByte()
	switch c {
	case '[': // CHARSET
		cs := p.scanCharset()
		frag = Literal(cs)
	case '.': // = [^\n]
		frag = Literal(rnSet{{'\n', '\n'}}.Canon(true))
	case '{': // NAME
		name := p.scanIdent()
		var ok bool
		frag, ok = p.defs[name]
		if ok {
			frag = frag.Clone()
		} else {
			p.errorf("undefined %q", name)
			frag = Epsilon()
		}
		if c, _ := p.rd.ReadByte(); c != '}' {
			p.error("missing '}'")
			p.rd.UnreadByte()
		}
	case '"': // LITERAL
		frag = p.parseLiteral()
	case '(':
		frag = p.parseRE()
		if c, _ := p.rd.ReadByte(); c != ')' {
			p.error("missing ')'")
			p.rd.UnreadByte()
		}
	default:
		p.rd.UnreadByte()
		r, _ := p.scanRune()
		frag = Literal(rnSet{{r, r}})
	}
L:
	for {
		c, _ := p.rd.ReadByte()
		switch c {
		case '*':
			frag = frag.Kleene(false)
		case '+':
			frag = frag.Kleene(true)
		default:
			p.rd.UnreadByte()
			break L
		}
	}
	return frag
}

func (p *parser) parseLiteral() frag {
	frags := []frag{}
L:
	for {
		r, escaped := p.scanRune()
		if !escaped {
			switch r {
			case '\n':
				p.rd.UnreadRune()
				p.error("missing '\"'")
				fallthrough
			case '"':
				break L
			}
		}
		frags = append(frags, Literal(rnSet{{r, r}}))
	}
	return Concat(frags)
}

func (p *parser) scanCharset() rnSet {
	rs := rnSet{}
	invert := false
	state := 0
	/* 0: - => character
	 * 1: - => dash (maybe)
	 * 2: just consumed a dash
	 */
	c, _ := p.rd.ReadByte()
	switch c {
	case ']':
		rs = append(rs, rn{']', ']'})
	case '^':
		invert = true
	default:
		p.rd.UnreadByte()
	}
L:
	for {
		r, escaped := p.scanRune()
		if !escaped {
			switch r {
			case '\n':
				p.error("missing '\"'")
				fallthrough
			case ']':
				break L
			case '-':
				if state == 1 {
					state = 2
					continue L
				}
			}
		}
		switch state {
		case 0, 1:
			rs = append(rs, rn{r, r})
			state = 1
		case 2:
			rs[len(rs)-1][1] = r
			state = 0
		}
	}
	if state == 2 {
		rs = append(rs, rn{'-', '-'})
	}
	if invert && len(rs) == 0 {
		invert = false
		rs = append(rs, rn{'^', '^'})
	}
	// fmt.Fprintf(os.Stderr, "rs = %+v\n", rs)
	return rs.Canon(invert)
}

func (p *parser) scanRune() (r rune, escaped bool) {
	r, _, _ = p.rd.ReadRune()
	if r == '\\' {
		escaped = true
		r, _, _ = p.rd.ReadRune()
		switch r {
		case 'a':
			r = '\a'
		case 'b':
			r = '\b'
		case 'f':
			r = '\f'
		case 'n':
			r = '\n'
		case 'r':
			r = '\r'
		case 't':
			r = '\t'
		case 'v':
			r = '\v'
		case 'o':
			r = p.scanDigits(r, 8, 3)
		case 'x':
			r = p.scanDigits(r, 16, 2)
		case 'u':
			r = p.scanDigits(r, 16, 4)
		case 'U':
			r = p.scanDigits(r, 16, 8)
		}
	}
	return
}

func (p *parser) scanDigits(r rune, base, length int) rune {
	s, _ := p.rd.Peek(length)
	if len(s) != length {
		return r
	}
	n, err := strconv.ParseUint(string(s), base, 32)
	if err != nil || n > unicode.MaxRune {
		return r
	}
	p.rd.Discard(length)
	return rune(n)
}
