package main

import ()

type fragState struct {
	Input rnSet
	Succ  []*fragState
}

func (s fragState) IsEpsilon() bool { return s.Input == nil }

// NFA fragment for construction
type frag struct {
	Start, Accept *fragState
}

func Epsilon() frag {
	return Literal(nil)
}

func Literal(rs rnSet) frag {
	s := &fragState{Input: rs}
	return frag{s, s}
}

func Concat(l []frag) frag {
	switch len(l) {
	case 0:
		return Epsilon()
	}
	for i, next := range l[1:] {
		l[i].Accept.Succ = append(l[i].Accept.Succ, next.Start)
	}
	return frag{l[0].Start, l[len(l)-1].Accept}
}

func Alter(l []frag) frag {
	switch len(l) {
	case 0:
		panic("nothing to alter")
	case 1:
		return l[0]
	}
	start := &fragState{}
	accept := &fragState{}
	for _, f := range l {
		start.Succ = append(start.Succ, f.Start)
		f.Accept.Succ = append(f.Accept.Succ, accept)
	}
	return frag{start, accept}
}

func (a frag) Kleene(plus bool) frag {
	ret := frag{
		Start: &fragState{Succ: []*fragState{a.Start}},
	}
	a.Accept.Succ = append(a.Accept.Succ, ret.Start)
	if plus {
		ret.Accept = a.Accept
	} else {
		ret.Accept = ret.Start
	}
	return ret
}

func (a frag) States() ([]*fragState, map[*fragState]int) {
	states := []*fragState{}
	ma := map[*fragState]int{}
	// DFS
	stack := []*fragState{a.Start}
	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		ma[s] = len(states)
		states = append(states, s)
		for _, succ := range s.Succ {
			if _, ok := ma[succ]; !ok {
				stack = append(stack, succ)
			}
		}
	}
	return states, ma
}

func (a frag) Clone() frag {
	states, ma := a.States()
	states1 := make([]*fragState, len(states))
	// create new states
	for i, s := range states {
		states1[i] = &fragState{Input: s.Input}
	}
	// fill in succ pointer
	for i, s := range states1 {
		for _, succ := range states[i].Succ {
			s.Succ = append(s.Succ, states1[ma[succ]])
		}
	}
	ret := frag{
		Start:  states1[ma[a.Start]],
		Accept: states1[ma[a.Accept]],
	}
	return ret
}

/*
func (a frag) Dump(w io.Writer) {
	states, ma := a.States()
	for i, s := range states {
		info := ""
		if s == a.Start {
			info = " (start)"
		} else if s == a.Accept {
			info = " (accept)"
		}
		succ := make([]int, len(s.Succ))
		for j, t := range s.Succ {
			succ[j] = ma[t]
		}
		fmt.Fprintf(w, "state %d%s, input = %v, succ = %v\n", i, info, s.Input, succ)
	}
}
*/
