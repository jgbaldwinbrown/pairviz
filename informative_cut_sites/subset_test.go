package main

import (
	"testing"
	"strings"
	"fmt"
)

const indata = `D00550:546:CD45VANXX:8:1101:1280:26822	2R_ISO1	18264965	2R_ISO1	18265308	+	-	UU
D00550:546:CD45VANXX:8:1101:1498:66931	mito_A4	160	mito_A4	18050	-	+	UU
D00550:546:CD45VANXX:8:1101:1524:76119	3L_A4	10002911	3L_A4	10003291	+	-	UU
D00550:546:CD45VANXX:8:1101:1853:94667	X_ISO1	21485453	X_ISO1	21485748	+	-	UU
D00550:546:CD45VANXX:8:1101:2091:96945	2L_ISO1	4956784	2L_ISO1	4957289	+	-	UU
D00550:546:CD45VANXX:8:1101:2334:28122	mito_A4	784	mito_A4	1052	+	-	UU
D00550:546:CD45VANXX:8:1101:2364:68366	2R_A4	12097393	2R_A4	12098181	+	-	UU
D00550:546:CD45VANXX:8:1101:2553:48904	3R_A4	27613446	3R_A4	27613786	+	-	UU
D00550:546:CD45VANXX:8:1101:2589:54221	3R_A4	14330626	3R_A4	14330874	+	-	UU
D00550:546:CD45VANXX:8:1101:2631:27774	3L_A4	11830606	3L_A4	11831227	+	-	UU
`

const cutbed = `3R_A4	14330625	14330625
3R_A4	14330873	14330873
`

const expect = `D00550:546:CD45VANXX:8:1101:2589:54221	3R_A4	14330626	3R_A4	14330874	+	-	UU
`

func TestFull(t *testing.T) {
	cutr := strings.NewReader(cutbed)
	_, cutsites, e := ReadCutsiteBedReader(cutr)
	if e != nil { panic(e) }
	fmt.Println(cutsites)

	r := strings.NewReader(indata)

	var w strings.Builder

	e = FilterPairviz(cutsites, r, &w)
	if e != nil { panic(e) }

	out := w.String()

	if out != expect {
		t.Errorf("out\n%v\n!= expect\n%v", out, expect)
	}
}
