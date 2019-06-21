package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

// the state machine
type mach struct {
	States []machState
}

type machState struct {
	Edges map[int]rnSet // k=dest, v=edge
	Tag   int
}

func (s *machState) DumpIfs() string {
	w := &strings.Builder{}
	type edge struct {
		rn  rn
		dst int
	}
	edges := []edge{}
	for k, v := range s.Edges {
		for _, rn := range v {
			edges = append(edges, edge{rn, k})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].rn[0] < edges[j].rn[0]
	})

	var dump func(int, int, bool)
	dump = func(l, r int, afterElse bool) {
		switch r - l {
		case 0:
			panic("bug: empty range in DumpIf")
		case 1:
			e := edges[l]
			cond := make([]string, 0, 2)
			if l == 0 {
				cond = append(cond, fmt.Sprintf("%q <= yyc ", e.rn[0]))
			}
			if l+1 == len(edges) || e.rn[1]+1 < edges[l+1].rn[0] {
				cond = append(cond, fmt.Sprintf("yyc <= %q ", e.rn[1]))
			}
			jump := fmt.Sprintf("goto yyS%d\n", e.dst)
			if len(cond) > 0 {
				fmt.Fprintf(w, "if %s {\n%s}\n", strings.Join(cond, " && "), jump)
			} else {
				if afterElse {
					w.WriteByte('{')
				}
				w.WriteString(jump)
				if afterElse {
					w.WriteByte('}')
				}
			}
		default:
			m := l + (r-l)/2
			e := edges[m]
			fmt.Fprintf(w, "if yyc < %q {\n", e.rn[0])
			dump(l, m, false)
			w.WriteString("} else\n")
			dump(m, r, true)
		}
	}
	if n := len(edges); n > 0 {
		dump(0, n, false)
	}
	return w.String()
}

func newMach(f frag, tags map[*fragState]int) mach {
	if f.Start.Input != nil {
		f.Start = &fragState{
			Succ: []*fragState{f.Start},
		}
	}
	states, ma := f.States()
	m := mach{
		States: make([]machState, len(states)),
	}
	for i, s := range states {
		edges := make(map[int]rnSet, len(s.Succ))
		for _, succ := range s.Succ {
			edges[ma[succ]] = succ.Input
		}
		tag, ok := tags[s]
		if !ok {
			tag = -1
		}
		m.States[i] = machState{
			Edges: edges,
			Tag:   tag,
		}
	}
	return m
}

//----------------------------------------------------------------------------
// NFA to DFA - power set construction

func (m mach) closures() []set {
	n := len(m.States)
	closure := make([]set, n)
	for i := range closure {
		closure[i] = set{}
		closure[i].Set(i)
	}
	for {
		changed := false
		for i, s := range m.States {
			m := closure[i]
			for dest, edge := range s.Edges {
				if edge == nil { // epsilon
					// m ← m ∪  closure[dest]
					for j := range closure[dest] {
						if _, ok := m[j]; !ok {
							changed = true
							m[j] = struct{}{}
						}
					}
				}
			}
		}
		if !changed {
			break
		}
	}
	return closure
}

func (m mach) dfa() mach {
	if len(m.States) == 0 {
		panic("machine has no states")
	}
	nClos := m.closures() // NFA id -> set of NFA id
	dClos := []set{}      // DFA id -> set of NFA id
	states := []machState{}
	ma := map[string]int{} // closure key -> state id
	makeState := func(clos set) int {
		key := clos.String()
		i, ok := ma[key]
		if !ok { // new state
			i = len(states)
			s := machState{
				Edges: make(map[int]rnSet),
				Tag:   -1,
			}
			states = append(states, s)
			dClos = append(dClos, clos)
			ma[key] = i
		}
		return i
	}

	makeState(nClos[0]) // start state

	// build other states
	for k := 0; k < len(states); k++ {
		s := &states[k]
		edges := []rnSet{} // outgoing range sets
		dests := []int{}   // outgoing destinations (NFA id)
		for i := range dClos[k] {
			if tag := m.States[i].Tag; s.Tag == -1 || tag != -1 && tag < s.Tag {
				s.Tag = tag
			}
			for dest, edge := range m.States[i].Edges {
				if edge != nil { // non-epsilon
					edges = append(edges, edge)
					dests = append(dests, dest)
				}
			}
		}
		rsFlat, rsMap := flattenRn(edges) // edges may overlap
		for i, ma := range rsMap {
			clos := set{}
			for j := range ma {
				clos.Union(nClos[dests[j]])
			}
			dest := makeState(clos)
			rn := rsFlat[i]
			s.Edges[dest] = append(s.Edges[dest], rn)
		}
		// by appending ranges to an edge, the edge may not
		// be canonical
		for dest, edge := range s.Edges {
			s.Edges[dest] = edge.Canon(false)
		}
	}
	return mach{
		States: states,
	}
}

//----------------------------------------------------------------------------
// DFA minimization - Hopcroft algorithm

func (m mach) invert() mach {
	states := make([]machState, len(m.States))
	for i := range states {
		states[i] = machState{
			Edges: map[int]rnSet{},
			Tag:   -1,
		}
	}
	for i, s := range m.States {
		states[i].Tag = s.Tag
		for dest, edge := range s.Edges {
			s := &states[dest]
			s.Edges[i] = append(s.Edges[i], edge...)
		}
	}
	for _, s := range states {
		for dest, edge := range s.Edges {
			s.Edges[dest] = edge.Canon(false)
		}
	}
	return mach{
		States: states,
	}
}

// data structure for partition refinement
type partitions struct {
	m    mach
	all  []int
	idx  []int
	part []int
	l, r []int // P[p] == all[l[p]:r[p]]
}

func (P *partitions) Init(m mach) {
	P.m = m
	n := len(m.States)

	P.all = make([]int, n)
	P.idx = make([]int, n)
	P.part = make([]int, n)
	for i := 0; i < n; i++ {
		P.all[i] = i
		P.idx[i] = i
	}
	sort.Sort(P)

	// create partitions - each tag is a distinct partition
	for i := range P.all {
		if i == 0 || P.Tag(i) != P.Tag(i-1) {
			P.l = append(P.l, i)
			P.r = append(P.r, i)
		}
		p := P.NPart() - 1
		P.part[P.all[i]] = p
		P.r[p]++
	}

}

func (P *partitions) Part(i int) []int {
	return P.all[P.l[i]:P.r[i]]
}

func (P *partitions) NPart() int {
	return len(P.l)
}

func (P *partitions) Len() int {
	return len(P.all)
}

func (P *partitions) Swap(i, j int) {
	x, y := P.all[i], P.all[j]
	P.all[i], P.all[j] = y, x
	P.idx[x], P.idx[y] = j, i
	P.part[x], P.part[y] = P.part[y], P.part[x]
}

func (P *partitions) Tag(i int) int {
	return P.m.States[P.all[i]].Tag
}

func (P *partitions) Less(i, j int) bool {
	return P.Tag(i) < P.Tag(j)
}

func (P *partitions) Print() {
	for i := range P.l {
		fmt.Fprintf(os.Stderr, "part %d: %v\n", i, P.all[P.l[i]:P.r[i]])
	}
	// for i := range P.m.States {
	// 	fmt.Printf("state %d is in part %d\n", i, P.part[i])
	// }
}

type partSort partitions

func (P *partSort) Len() int {
	return len(P.l)
}

func (P *partSort) Less(i, j int) bool {
	return P.all[P.l[i]] < P.all[P.l[j]]
}

func (P *partSort) Swap(i, j int) {
	for _, s := range P.all[P.l[i]:P.r[i]] {
		P.part[s] = j
	}
	for _, s := range P.all[P.l[j]:P.r[j]] {
		P.part[s] = i
	}
	P.l[i], P.l[j] = P.l[j], P.l[i]
	P.r[i], P.r[j] = P.r[j], P.r[i]
}

func (P *partitions) Sort() {
	for i := 0; i < P.NPart(); i++ {
		sort.Ints(P.Part(i))
	}
	sort.Sort((*partSort)(P))
}

func (m mach) minimize() mach {
	var P partitions

	P.Init(m)

	// calculate predecessors of each states
	inv := m.invert()

	// create worklist - initially all partitions except the first
	W := set{} // set of partition id
	for i := 0; i < P.NPart(); i++ {
		W.Set(i)
	}

	// do partition refinement until worklist is empty
	for len(W) > 0 {
		// P.Print()
		var A []int // pick a partition from W
		for p := range W {
			// println("pick", p)
			A = P.Part(p)
			W.Clear(p)
			break
		}

		// get all transitions from some state to any state in A
		trans := map[int]rnSet{}
		for _, s := range A {
			for dest, edge := range inv.States[s].Edges {
				trans[dest] = append(trans[dest], edge...)
			}
		}
		// then flatten the ranges and iterate through each range
		edges := []rnSet{}
		preds := []int{}
		for k, v := range trans {
			edges = append(edges, v.Canon(false))
			preds = append(preds, k)
		}
		_, rsMap := flattenRn(edges)
		for _, ma := range rsMap {
			X := map[int][]int{} // collect predecessors by partition
			for j := range ma {
				pred := preds[j]
				p := P.part[pred]
				X[p] = append(X[p], pred)
			}
			// fmt.Fprintln(os.Stderr, "X =", X)
			for p, x := range X {
				len_x := len(x)
				len_y := P.r[p] - P.l[p]
				if len_x == len_y {
					continue
				}
				q := P.NPart()
				P.l = append(P.l, P.r[p])
				P.r = append(P.r, P.r[p])
				for _, s := range x {
					// move s from p (Y) to q (new partition)
					P.r[p]--
					P.l[q]--
					P.Swap(P.idx[s], P.l[q])
					P.part[s] = q
				}
				if W.Test(p) || len_x < len_y {
					W.Set(q)
				} else {
					W.Set(p)
				}
			}
		}
	}

	P.Sort()
	//P.Print()

	// construct a new DFA
	states := make([]machState, P.NPart())
	for i := range states {
		s := machState{
			Edges: map[int]rnSet{},
			Tag:   -1,
		}
		for _, j := range P.Part(i) {
			for k, v := range m.States[j].Edges {
				k1 := P.part[k]
				s.Edges[k1] = append(s.Edges[k1], v...)
			}
			if tag := m.States[j].Tag; s.Tag == -1 {
				s.Tag = tag
			} else if s.Tag != tag {
				panic("bug: minimization coalesced states with different tags")
			}
		}
		for k, v := range s.Edges {
			s.Edges[k] = v.Canon(false)
		}
		states[i] = s
	}

	return mach{states}
}

// output machine in dot format
var dotTmpl = template.Must(template.New("machdot").Parse(`
digraph mach {
	rankdir = LR;
	node[shape=circle];
{{- range $i, $s := .States }}
	n{{$i}} [label="{{$i}}"{{if ne $s.Tag -1}},xlabel="{{$s.Tag}}",peripheries=2{{end}}];
{{- range $j, $r := $s.Edges}}
	n{{$i}}->n{{$j}} [label={{if $r}}{{printf "%q" $r}}{{else}}ϵ{{end}}];
{{- end}}
{{- end}}
}
`))
