package main

import (
	"testing"
	"strings"
)

func TestFprintHits(t *testing.T) {
	counts := map[Pos]int64 {
		Pos{"chr1", 0}: 5,
		Pos{"chr1", 1}: 0,
	}

	poses := []Pos{Pos{"chr1", 0}, Pos{"chr1", 1}, Pos{"chr1", 2}}

	var b strings.Builder
	_, e := FprintHits(&b, poses, counts)
	if e != nil { panic(e) }

	expect := `chr1	0	0	5
chr1	1	1	0
chr1	2	2	0
`

	out := b.String()
	if out != expect {
		t.Errorf("out != expect\nout:\n%v\nexpect:\n%v\n", out, expect)
	}
}
