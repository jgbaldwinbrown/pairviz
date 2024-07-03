package prepfa

import (
	"fmt"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"testing"
)

var infa = []fastats.FaEntry {
	fastats.FaEntry {
		Header: "1:2-5",
		Seq: "at",
	},
	fastats.FaEntry {
		Header: "1:5-8",
		Seq: "gc",
	},
	fastats.FaEntry {
		Header: "1:8-11",
		Seq: "tt",
	},
}

var inbedshort = []fastats.BedEntry[[]string] {
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{3, 6},
	}},
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{6, 9},
	}},
}

var inbedlong = []fastats.BedEntry[[]string] {
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{3, 6},
	}},
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{6, 9},
	}},
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{9, 12},
	}},
	fastats.BedEntry[[]string]{ChrSpan: fastats.ChrSpan{
		Chr: "1",
		Span: fastats.Span{12, 15},
	}},
}

func TestCleanupShortBed(t *testing.T) {
	set, e := FaChrSpanSet(AddErr(SliceIter(infa)))
	if e != nil {
		panic(e)
	}
	fmt.Println("set:", set)

	bedkept, e := CollectErr(KeepBedMatches[[]string](set, AddErr(SliceIter(inbedshort))))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](AddErr(SliceIter(bedkept)))
	fmt.Println("bedset:", bedset)
	fa1kept, e := CollectErr(KeepFaMatches(bedset, AddErr(SliceIter(infa))))
	if e != nil {
		panic(e)
	}

	if len(bedkept) != len(fa1kept) {
		t.Errorf("len(bedkept) %v != len(fa1kept) %v", len(bedkept), len(fa1kept))
	}

	fmt.Printf("bedkept: %v\nfakept: %v\n", bedkept, fa1kept)
}

func TestCleanupLongBed(t *testing.T) {
	set, e := FaChrSpanSet(AddErr(SliceIter(infa)))
	if e != nil {
		panic(e)
	}
	fmt.Println("set:", set)

	bedkept, e := CollectErr(KeepBedMatches[[]string](set, AddErr(SliceIter(inbedlong))))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](AddErr(SliceIter(bedkept)))
	fmt.Println("bedset:", bedset)
	fa1kept, e := CollectErr(KeepFaMatches(bedset, AddErr(SliceIter(infa))))
	if e != nil {
		panic(e)
	}

	if len(bedkept) != len(fa1kept) {
		t.Errorf("len(bedkept) %v != len(fa1kept) %v", len(bedkept), len(fa1kept))
	}

	fmt.Printf("bedkept: %v\nfakept: %v\n", bedkept, fa1kept)
}
