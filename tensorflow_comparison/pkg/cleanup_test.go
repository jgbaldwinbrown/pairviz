package prepfa

import (
	"fmt"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/iter"
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
	set, e := FaChrSpanSet(iter.SliceIter[fastats.FaEntry](infa))
	if e != nil {
		panic(e)
	}
	fmt.Println("set:", set)

	bedkept, e := iter.Collect[fastats.BedEntry[[]string]](KeepBedMatches[[]string](set, iter.SliceIter[fastats.BedEntry[[]string]](inbedshort)))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](iter.SliceIter[fastats.BedEntry[[]string]](bedkept))
	fmt.Println("bedset:", bedset)
	fa1kept, e := iter.Collect[fastats.FaEntry](KeepFaMatches(bedset, iter.SliceIter[fastats.FaEntry](infa)))
	if e != nil {
		panic(e)
	}

	if len(bedkept) != len(fa1kept) {
		t.Errorf("len(bedkept) %v != len(fa1kept) %v", len(bedkept), len(fa1kept))
	}

	fmt.Printf("bedkept: %v\nfakept: %v\n", bedkept, fa1kept)
}

func TestCleanupLongBed(t *testing.T) {
	set, e := FaChrSpanSet(iter.SliceIter[fastats.FaEntry](infa))
	if e != nil {
		panic(e)
	}
	fmt.Println("set:", set)

	bedkept, e := iter.Collect[fastats.BedEntry[[]string]](KeepBedMatches[[]string](set, iter.SliceIter[fastats.BedEntry[[]string]](inbedlong)))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](iter.SliceIter[fastats.BedEntry[[]string]](bedkept))
	fmt.Println("bedset:", bedset)
	fa1kept, e := iter.Collect[fastats.FaEntry](KeepFaMatches(bedset, iter.SliceIter[fastats.FaEntry](infa)))
	if e != nil {
		panic(e)
	}

	if len(bedkept) != len(fa1kept) {
		t.Errorf("len(bedkept) %v != len(fa1kept) %v", len(bedkept), len(fa1kept))
	}

	fmt.Printf("bedkept: %v\nfakept: %v\n", bedkept, fa1kept)
}
