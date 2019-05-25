package main

import (
	"sort"
	"strconv"
	"unicode"
)

type rn [2]rune // Unicode range
type rnSet []rn // set of ranges

func Char(r rune) rnSet {
	return rnSet{{r, r}}
}

// append a single rune (special character escaped) to byte slice
func appendRn(buf []byte, r rune) []byte {
	switch r {
	case '-', '^', ']': // need escape inside []
		buf = append(buf, '\\', byte(r))
		return buf
	default:
		buf2 := strconv.AppendQuoteRune(make([]byte, 0, 16), r)
		// strip the single quotes
		return append(buf, buf2[1:len(buf2)-1]...)
	}
}

// append range into byte slice in proper format
func (r rn) appendTo(buf []byte) []byte {
	buf = appendRn(buf, r[0])
	switch n := r[1] - r[0]; n {
	default:
		buf = append(buf, '-')
		fallthrough
	case 1:
		buf = appendRn(buf, r[1])
	case 0:
	}
	return buf
}

func (r rn) String() string {
	return string(r.appendTo(nil))
}

// implement sort.Interface, ordered by left end
func (s rnSet) Len() int           { return len(s) }
func (s rnSet) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s rnSet) Less(i, j int) bool { return s[i][0] < s[j][0] }

func (s rnSet) String() string {
	if s == nil {
		return "ϵ"
	}
	buf := []byte{'['}
	if n := len(s); n > 0 && s[n-1][1] == unicode.MaxRune && s[n-1][0] != 0 {
		// probably an inverted range
		buf = append(buf, '^')
		s = append(rnSet{}, s...).Canon(true)
	}
	for _, r := range s {
		buf = r.appendTo(buf)
	}
	buf = append(buf, ']')
	return string(buf)
}

// rearrange ranges so that they
// 1) are separate (disjoint and not adjacent), and
// 2) appear in increasing order.
// The original set may be altered.
func (s rnSet) Canon(invert bool) (ret rnSet) {
	// linear time other than sorting
	sort.Sort(s)
	curr := rn{0, -1}
	ret = s[:0] // we can actually reuse the space
	for _, r := range s {
		if r[0] > curr[1]+1 {
			if invert {
				curr = rn{curr[1] + 1, r[0] - 1}
			}
			if curr[0] <= curr[1] {
				ret = append(ret, curr)
			}
			curr = r
		} else if r[1] > curr[1] {
			curr[1] = r[1]
		}
	}
	if invert {
		curr = rn{curr[1] + 1, unicode.MaxRune}
	}
	if curr[0] <= curr[1] {
		ret = append(ret, curr)
	}
	return
}

// for flattening
type rnEv struct {
	id int
	r  rune
}

type rnEvs []rnEv

func (s rnEvs) Len() int           { return len(s) }
func (s rnEvs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s rnEvs) Less(i, j int) bool { return s[i].r < s[j].r }

// rearrange canonicalized range sets into disjoint ranges,
// each mapping to a set of sets.
func flattenRn(ss []rnSet) (ret []rn, ma []set) {
	var ev rnEvs
	for id, s := range ss {
		for _, r := range s {
			ev = append(ev, rnEv{id, r[0]}, rnEv{id, r[1] + 1})
		}
	}
	sort.Sort(ev)
	s := set{}
	curr := rune(0)
	for i, e := range ev {
		if i == 0 || ev.Less(i-1, i) {
			if len(s) > 0 {
				ret = append(ret, rn{curr, e.r - 1})
				ma = append(ma, s.Clone())
			}
			curr = e.r
		}
		s.Flip(e.id)
	}
	return
}
