package pairviz

import (
	"testing"
	"math"
)

type FpkmArgs struct {
	Name string
	Count int64
	Total int64
	Len int64
	Fpkm float64
}

func GetAbsBig(x, y float64) (big, small float64) {
	if math.Abs(x) > math.Abs(y) {
		return x, y
	}
	return y, x
}

func AeqOrBothNan(x, y float64) bool {
	if math.IsNaN(x) && math.IsNaN(y) {
		return true
	}
	if x == y {
		return true
	}

	big, small := GetAbsBig(x, y)
	thresh := big / 1e10
	return math.Abs(big - small) < thresh
}

func TestFpkm(t *testing.T) {
	tests := []FpkmArgs {
		FpkmArgs{"good", 25, 1000, 100, 250000},
		FpkmArgs{"zerolen", 25, 1000, 0, math.Inf(1)},
		FpkmArgs{"zerototal", 25, 0, 100, math.Inf(1)},
		FpkmArgs{"zerocount", 0, 1000, 100, 0},
		FpkmArgs{"zerocounttotal", 0, 0, 100, math.NaN()},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fpkm := Fpkm(test.Count, test.Total, test.Len)
			if !AeqOrBothNan(fpkm, test.Fpkm) {
				t.Errorf("fpkm %v != test.Fpkm %v", fpkm, test.Fpkm)
			}
		})
	}
}
