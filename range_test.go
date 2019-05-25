package main

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	/* ABCDEFGHIJKLMNOPQR
	 * ---------------    1
	 *    ----  ---  - -- 2
	 *    --------  --- - 4
	 * 111777755773157426 */
	s0 := []rnSet{
		{{'A', 'O'}},
		{{'D', 'G'}, {'J', 'L'}, {'O', 'O'}, {'Q', 'R'}},
		{{'D', 'K'}, {'N', 'P'}, {'R', 'R'}},
	}
	s1, _ := flattenRn(s0)

	expect := []rn{
		{'A', 'C'}, {'D', 'G'}, {'H', 'I'}, {'J', 'K'},
		{'L', 'L'}, {'M', 'M'}, {'N', 'N'}, {'O', 'O'},
		{'P', 'P'}, {'Q', 'Q'}, {'R', 'R'},
	}

	if !reflect.DeepEqual(s1, expect) {
		t.Errorf("s1 = %v, expect %v", s1, expect)
	}

	s0 = []rnSet{{{'a', 'a'}, {'c', 'c'}}}
	s1, _ = flattenRn(s0)

	expect = []rn{{'a', 'a'}, {'c', 'c'}}
	if !reflect.DeepEqual(s1, expect) {
		t.Errorf("s1 = %v, expect %v", s1, expect)
	}
}
