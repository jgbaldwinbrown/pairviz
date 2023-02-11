package main

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

func TestFpkm(t *testing.T) {
	tests := []FpkmArgs {
		FpkmArgs{"good", 25, 1000, 100, 250000},
		FpkmArgs{"good", 25, 1000, 0, math.NaN()},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fpkm := Fpkm(test.Count, test.Total, test.Len)
			if fpkm != test.Fpkm {
				t.Errorf("fpkm %v != test.Fpkm %v", fpkm, test.Fpkm)
			}
		})
	}
}
