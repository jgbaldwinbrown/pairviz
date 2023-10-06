package pairviz

import (
	"fmt"
	"testing"
	"strings"
)

func TestEcNorm(t *testing.T) {
	flags := gFlags
	in := strings.NewReader(gTestIn)
	var b strings.Builder

	FprintWinStats(&b, WinStats(flags, in), flags.SeparateGenomes, flags.ReadLen, true)
	p := b.String()

	cchr := "2L"
	in = strings.NewReader(p)
	control, e := GetControlStatMeans(cchr, ParsePairvizOut(in))
	Must(e)
	fmt.Println(control)

	// in = strings.NewReader(p)
	// subit := SubtractControlStatAll(ParsePairvizOut(in), control)
	// subs, e := Collect[JsonOutStat](subit)
	// Must(e)
	// fmt.Println(subs)
}
