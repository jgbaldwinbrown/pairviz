package prepfa

import (
	"math"
	"strconv"
	"strings"
	"bufio"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/iter"
	"github.com/jgbaldwinbrown/csvh"
	"flag"
	"regexp"
	"sort"
	"fmt"
)

var parentRe = regexp.MustCompile(`_(.*)$`)

func FilterChrParent[T any](parent string, it iter.Iter[fastats.BedEntry[T]]) *iter.Iterator[fastats.BedEntry[T]] {
	return &iter.Iterator[fastats.BedEntry[T]]{Iteratef: func(yield func(fastats.BedEntry[T]) error) error {
		return it.Iterate(func(b fastats.BedEntry[T]) error {
			fields := parentRe.FindStringSubmatch(b.Chr)
			if fields == nil {
				return nil
			}
			if fields[1] != parent {
				return nil
			}
			return yield(b)
		})
	}}
}

func StripChrParent[T any](it iter.Iter[fastats.BedEntry[T]]) *iter.Iterator[fastats.BedEntry[T]] {
	return iter.Transform[fastats.BedEntry[T], fastats.BedEntry[T]](it, func(f fastats.BedEntry[T]) (fastats.BedEntry[T], error) {
		f.Chr = parentRe.ReplaceAllString(f.Chr, "")
		return f, nil
	})
}

var faPosRe = regexp.MustCompile(`^([^:]*):([^-]*)-(.*)$`)

func ChrSpanLess(cs1, cs2 fastats.ChrSpan) bool {
	if cs1.Chr < cs2.Chr {
		return true
	} else if cs1.Chr > cs2.Chr {
		return false
	} else if cs1.Start < cs2.Start {
		return true
	} else if cs1.Start > cs2.Start {
		return false
	} else {
		return cs1.End < cs2.End
	}
}

func ChrSpanEq(cs1, cs2 fastats.ChrSpan) bool {
	return cs1.Chr == cs2.Chr &&
		cs1.Start == cs2.Start &&
		cs1.End == cs2.End
	if cs1.Chr < cs2.Chr {
		return true
	} else if cs1.Chr > cs2.Chr {
		return false
	} else if cs1.Start < cs2.Start {
		return true
	} else if cs1.Start > cs2.Start {
		return false
	} else {
		return cs1.End < cs2.End
	}
}

func GetFaChrSpan(f fastats.FaEntry) (fastats.ChrSpan, error) {
	fields := faPosRe.FindStringSubmatch(f.Header)

	var cs fastats.ChrSpan
	_, e := csvh.Scan(fields[1:], &cs.Chr, &cs.Start, &cs.End)
	return cs, e
}

func FaPosLess(f1, f2 fastats.FaEntry) bool {
	cs1, e := GetFaChrSpan(f1)
	if e != nil {
		panic(e)
	}
	cs2, e := GetFaChrSpan(f2)
	if e != nil {
		panic(e)
	}

	return ChrSpanLess(cs1, cs2)
}

func SortFa(fa []fastats.FaEntry) {
	sort.Slice(fa, func(i, j int) bool {
		return FaPosLess(fa[i], fa[j])
	})
}

func SortBed[T any](bed []fastats.BedEntry[T]) {
	sort.Slice(bed, func(i, j int) bool {
		return ChrSpanLess(bed[i].ChrSpan, bed[j].ChrSpan)
	})
}

func KeepBedMatches[T any](css map[string]int, it iter.Iter[fastats.BedEntry[T]]) *iter.Iterator[fastats.BedEntry[T]] {
	counts := map[string]int{}
	for chr, _ := range css {
		counts[chr] = 0
	}

	return &iter.Iterator[fastats.BedEntry[T]]{Iteratef: func(yield func(fastats.BedEntry[T]) error) error {
		return it.Iterate(func(b fastats.BedEntry[T]) error {
			size := b.End - b.Start
			if b.Start % size != 0 {
				return nil
			}

			if count, ok := counts[b.Chr]; ok {
				// log.Printf("ok\n")
				oknum := count < css[b.Chr]
				counts[b.Chr]++
				// log.Printf("count: %v; css[b.Chr]: %v; oknum: %v; b: %v\n", count, css[b.Chr], oknum, b)
				if oknum {
					return yield(b)
				}
			}
			return nil
		})
	}}
}

func KeepFaMatches(css map[string]int, it iter.Iter[fastats.FaEntry]) *iter.Iterator[fastats.FaEntry] {
	counts := map[string]int{}
	for chr, _ := range css {
		counts[chr] = 0
	}

	return &iter.Iterator[fastats.FaEntry]{Iteratef: func(yield func(fastats.FaEntry) error) error {
		return it.Iterate(func(b fastats.FaEntry) error {
			facs, err := GetFaChrSpan(b)
			if err != nil {
				return err
			}
			fachr := facs.Chr

			if count, ok := counts[fachr]; ok {
				// log.Printf("ok\n")
				oknum := count < css[fachr]
				counts[fachr]++
				// log.Printf("count: %v; css[fachr]: %v; oknum: %v; b: %v\n", count, css[fachr], oknum, b)
				if oknum {
					return yield(b)
				}
			}
			return nil
		})
	}}
}

func CollectBedWithHeader(path string) ([]fastats.BedEntry[[]string], error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	br := bufio.NewReader(r)

	_, err := br.ReadString('\n')

	bed, err := iter.Collect[fastats.BedEntry[[]string]](fastats.ParseBed[[]string](br, func(fields []string) ([]string, error) {
		out := make([]string, len(fields))
		copy(out, fields)
		return out, nil
	}))
	return bed, err
}

func FaChrSpanSet(it iter.Iter[fastats.FaEntry]) (map[string]int, error) {
	m := map[string]int{}
	err := it.Iterate(func(f fastats.FaEntry) error {
		fields := faPosRe.FindStringSubmatch(f.Header)
		m[fields[1]]++
		return nil
	})

	return m, err
}

func BedSet[T any](it iter.Iter[fastats.BedEntry[T]]) map[string]int {
	m := map[string]int{}
	_ = it.Iterate(func(f fastats.BedEntry[T]) error {
		m[f.Chr]++
		return nil
	})

	return m
}

func WriteBed(path string, bed iter.Iter[fastats.BedEntry[[]string]]) error {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	return bed.Iterate(func(f fastats.BedEntry[[]string]) error {
		_, e := fmt.Fprintf(bw, "%v\t%v\t%v\t%v\n", f.Chr, f.Start, f.End, strings.Join(f.Fields, "\t"))
		return e
	})
}

type CleanupFlags struct {
	Fa1 string
	Fa2 string
	Bed string
	Outpre string
	Parent string
	Paircol int
}

func GetCleanupFlags() (CleanupFlags, error) {
	var f CleanupFlags

	flag.StringVar(&f.Fa1, "fa1", "", "path to .fa or .fa.gz file")
	flag.StringVar(&f.Fa2, "fa2", "", "path to .fa or .fa.gz file")
	flag.StringVar(&f.Bed, "bed", "", "path to .bed or .bed.gz file")
	flag.StringVar(&f.Outpre, "o", "out", "output prefix")
	flag.StringVar(&f.Parent, "p", "", "parent to keep")
	flag.IntVar(&f.Paircol, "c", 7, "Column in bed files specifying pairing rate")
	flag.Parse()

	if f.Fa1 == "" {
		panic(fmt.Errorf("mising -fa1"))
	}
	if f.Fa2 == "" {
		panic(fmt.Errorf("mising -fa2"))
	}
	if f.Bed == "" {
		panic(fmt.Errorf("missing -bed"))
	}
	if f.Parent == "" {
		panic(fmt.Errorf("missing -p"))
	}

	return f, nil
}

type StripNaNsOut struct {
	Bed []fastats.BedEntry[[]string]
	Fa1 []fastats.FaEntry
	Fa2 []fastats.FaEntry
}

func StripNaNs(col int, bed []fastats.BedEntry[[]string], fa1, fa2 []fastats.FaEntry) (StripNaNsOut, error) {
	out := StripNaNsOut {
		Bed: make([]fastats.BedEntry[[]string], 0, len(bed)),
		Fa1: make([]fastats.FaEntry, 0, len(fa1)),
		Fa2: make([]fastats.FaEntry, 0, len(fa2)),
	}

	if len(bed) != len(fa1) || len(bed) != len(fa2) {
		return out, fmt.Errorf("StripNaNs: lengths don't match; len(bed) %v, len(fa1) %v, len(fa2) %v", len(bed), len(fa1), len(fa2))
	}

	for i, _ := range bed {
		f, err := strconv.ParseFloat(bed[i].Fields[col], 64)
		if err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			out.Bed = append(out.Bed, bed[i])
			out.Fa1 = append(out.Fa1, fa1[i])
			out.Fa2 = append(out.Fa2, fa2[i])
		}
	}

	return out, nil
}

func Cleanup(f CleanupFlags) error {
	fa1, e := CollectFa(f.Fa1)
	if e != nil {
		panic(e)
	}
	SortFa(fa1)

	fa2, e := CollectFa(f.Fa2)
	if e != nil {
		panic(e)
	}
	if e != nil {
		panic(e)
	}
	SortFa(fa2)

	bed, e := CollectBedWithHeader(f.Bed)
	if e != nil {
		panic(e)
	}

	pbed := FilterChrParent[[]string](f.Parent, iter.SliceIter[fastats.BedEntry[[]string]](bed))
	bed, e = iter.Collect[fastats.BedEntry[[]string]](StripChrParent[[]string](pbed))
	if e != nil {
		panic(e)
	}

	set, e := FaChrSpanSet(iter.SliceIter[fastats.FaEntry](fa1))
	if e != nil {
		panic(e)
	}
	SortBed(bed)
	// log.Println("set:", set)

	bedkept, e := iter.Collect[fastats.BedEntry[[]string]](KeepBedMatches[[]string](set, iter.SliceIter[fastats.BedEntry[[]string]](bed)))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](iter.SliceIter[fastats.BedEntry[[]string]](bedkept))
	fa1kept, e := iter.Collect[fastats.FaEntry](KeepFaMatches(bedset, iter.SliceIter[fastats.FaEntry](fa1)))
	if e != nil {
		panic(e)
	}
	fa2kept, e := iter.Collect[fastats.FaEntry](KeepFaMatches(bedset, iter.SliceIter[fastats.FaEntry](fa2)))
	if e != nil {
		panic(e)
	}

	stripped, e := StripNaNs(
		f.Paircol - 3,
		bedkept,
		fa1kept,
		fa2kept,
	)
	if e != nil {
		panic(e)
	}

	if e := WriteBed(f.Outpre + ".bed.gz", iter.SliceIter[fastats.BedEntry[[]string]](stripped.Bed)); e != nil {
		panic(e)
	}
	if e := WriteFasta(f.Outpre + "_1.fa.gz", iter.SliceIter[fastats.FaEntry](stripped.Fa1)); e != nil {
		panic(e)
	}
	if e := WriteFasta(f.Outpre + "_2.fa.gz", iter.SliceIter[fastats.FaEntry](stripped.Fa2)); e != nil {
		panic(e)
	}

	return nil
}

func RunCleanup() {
	flags, e := GetCleanupFlags()
	if e != nil {
		panic(e)
	}
	e = Cleanup(flags)
	if e != nil {
		panic(e)
	}
}
