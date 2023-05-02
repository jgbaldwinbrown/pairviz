package register

import (
	"testing"
	"strings"
)

const filterArg = `
{
	"FilterSets": [
		{
			"MaxDist": 5,
			"MinDist": 2,
			"Faces": [1,2,3],
			"ReadTypes": [1,2,3]
		}
	]
}
`

const filterIn = `r1	c1_p1	3	c1_p1	8	+	-	UU
r2	c1_p1	3	c1_p2	6	+	-	UU
r3	c1_p1	1	c2_p1	9	+	-	UU
r3	c1_p1	1	c2_p2	2	+	-	UU`

const filterExpect = `r1	c1_p1	3	c1_p1	8	+	-	UU
r2	c1_p1	3	c1_p2	6	+	-	UU
`

func TestFilter(t *testing.T) {
	arg, e := GetFilterArgsFromReader(strings.NewReader(filterArg))
	if e != nil { panic(e) }

	var b strings.Builder
	e = RunFilter(strings.NewReader(filterIn), &b, arg)
	if e != nil { panic(e) }

	out := b.String()
	if out != filterExpect {
		t.Errorf("out %v != expect %v", out, filterExpect)
	}
}
