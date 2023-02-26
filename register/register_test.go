package main

import (
	"strings"
	"testing"
)

const in = `first	3L_ISO1	1	3L_ISO1	2	-	+	UU
second	3L_ISO1	5	3L_ISO1	8	-	+	UU
second	3L_ISO1	6	3L_ISO1	9	-	+	UU
second	3L_ISO1	5	3L_W501	10	-	+	UU
second	3R_ISO1	5	3L_W501	20	-	+	UU
`

const in2 = `first	3L_ISO1	1	3L_ISO1	100001	-	+	UU
`

const expect = `0	0	0	0
1	0	1	0
2	0	0	0
3	0	2	0
4	0	0	0
5	1	0	0
6	0	0	0
7	0	0	0
8	0	0	0
9	0	0	0
10	0	0	0
11	0	0	0
12	0	0	0
13	0	0	0
14	0	0	0
15	0	0	1
`

const inreal = `A00421:241:HMKCMDRXX:2:1101:1000:1344	3L_ISO1	23059614	3L_ISO1	23639308	-	+	UU
A00421:241:HMKCMDRXX:2:1101:1000:1470	X_W501	19252389	X_W501	19254964	+	+	UU
A00421:241:HMKCMDRXX:2:1101:1000:1626	2R_ISO1	14374650	2R_ISO1	14404359	+	-	UU
A00421:241:HMKCMDRXX:2:1101:1000:2440	X_W501	7327312	X_W501	7327506	+	-	UU
A00421:241:HMKCMDRXX:2:1101:1000:5102	2R_ISO1	9566998	2R_ISO1	9567133	-	+	UU
A00421:241:HMKCMDRXX:2:1101:1000:5572	X_ISO1	8645922	X_ISO1	8649171	-	-	UU
A00421:241:HMKCMDRXX:2:1101:1000:5666	2R_ISO1	20575535	2R_ISO1	20596905	-	+	UU
A00421:241:HMKCMDRXX:2:1101:1000:6417	3L_ISO1	17213063	3L_W501	17105970	-	+	UU
A00421:241:HMKCMDRXX:2:1101:1000:7326	2L_ISO1	8052080	2L_W501	8078172	+	-	UU
A00421:241:HMKCMDRXX:2:1101:1000:7545	3R_W501	4476168	3R_W501	4486295	-	-	UU
`

func TestRun(t *testing.T) {
	r := strings.NewReader(in)
	var b strings.Builder
	e := Run(r, &b)
	if e != nil { panic(e) }

	out := b.String()
	if out != expect {
		t.Errorf("out %v != expect %v", out, expect)
	}
}

func TestRun2(t *testing.T) {
	r := strings.NewReader(in)
	var b strings.Builder
	e := Run(r, &b)
	if e != nil { panic(e) }

	out := b.String()
	if out != expect {
		t.Errorf("out %v != expect %v", out, expect)
	}
}
