package main

import (
	"sort"
	"strconv"
)

type set map[int]struct{}

func (s set) Set(i int) {
	s[i] = struct{}{}
}

func (s set) Clear(i int) {
	delete(s, i)
}

func (s set) Test(i int) bool {
	_, ok := s[i]
	return ok
}

func (s set) Flip(i int) {
	if s.Test(i) {
		s.Clear(i)
	} else {
		s.Set(i)
	}
}

func (s set) Clone() set {
	s1 := set{}
	for i := range s {
		s1.Set(i)
	}
	return s1
}

func (s set) Union(s1 set) (changed bool) {
	for i := range s1 {
		if !s.Test(i) {
			changed = true
			s.Set(i)
		}
	}
	return
}

func (s set) String() string {
	ints := make([]int, 0, len(s))
	for i := range s {
		ints = append(ints, i)
	}
	sort.Ints(ints)
	b := []byte{'{'}
	for i, n := range ints {
		if i > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendInt(b, int64(n), 10)
	}
	b = append(b, '}')
	return string(b)
}
