package prepfa

import (
	"math"
	"strconv"
	"strings"
	"bufio"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"iter"
	"github.com/jgbaldwinbrown/csvh"
	"flag"
	"regexp"
	"sort"
	"fmt"
)

func CollectErr[T any](it iter.Seq2[T, error]) ([]T, error) {
	var o []T
	for t, e := range it {
		if e != nil {
			return nil, e
		}
		o = append(o, t)
	}
	return o, nil
}

var parentRe = regexp.MustCompile(`_(.*)$`)

func FilterChrParent[T any](parent string, it iter.Seq2[fastats.BedEntry[T], error]) iter.Seq2[fastats.BedEntry[T], error] {
	return func(yield func(fastats.BedEntry[T], error) bool) {
		for b, e := range it {
			if e != nil {
				yield(fastats.BedEntry[T]{}, e)
				return
			}
			fields := parentRe.FindStringSubmatch(b.Chr)
			if fields == nil {
				continue
			}
			if fields[1] != parent {
				continue
			}
			if ok := yield(b, nil); !ok {
				return
			}
		}
	}
}

func StripChrParent[T any](it iter.Seq2[fastats.BedEntry[T], error]) iter.Seq2[fastats.BedEntry[T], error] {
	return func(yield func(fastats.BedEntry[T], error) bool) {
		for f, err := range it {
			if err != nil {
				yield(fastats.BedEntry[T]{}, err)
				return
			}
			f.Chr = parentRe.ReplaceAllString(f.Chr, "")
			if ok := yield(f, nil); !ok {
				return
			}
		}
	}
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

func KeepBedMatches[T any](css map[string]int, it iter.Seq2[fastats.BedEntry[T], error]) iter.Seq2[fastats.BedEntry[T], error] {
	counts := map[string]int{}
	for chr, _ := range css {
		counts[chr] = 0
	}

	return func(yield func(fastats.BedEntry[T], error) bool) {
		for b, err := range it {
			if err != nil {
				yield(fastats.BedEntry[T]{}, err)
				return
			}
			size := b.End - b.Start
			if b.Start % size != 0 {
				continue
			}

			if count, ok := counts[b.Chr]; ok {
				// log.Printf("ok\n")
				oknum := count < css[b.Chr]
				counts[b.Chr]++
				// log.Printf("count: %v; css[b.Chr]: %v; oknum: %v; b: %v\n", count, css[b.Chr], oknum, b)
				if oknum {
					if ok := yield(b, nil); !ok {
						return
					}
				}
			}
		}
	}
}

func KeepFaMatches(css map[string]int, it iter.Seq2[fastats.FaEntry, error]) iter.Seq2[fastats.FaEntry, error] {
	counts := map[string]int{}
	for chr, _ := range css {
		counts[chr] = 0
	}

	return func(yield func(fastats.FaEntry, error) bool) {
		for b, err := range it {
			if err != nil {
				yield(fastats.FaEntry{}, err)
				return
			}
			facs, err := GetFaChrSpan(b)
			if err != nil {
				yield(fastats.FaEntry{}, err)
				return
			}
			fachr := facs.Chr

			if count, ok := counts[fachr]; ok {
				// log.Printf("ok\n")
				oknum := count < css[fachr]
				counts[fachr]++
				// log.Printf("count: %v; css[fachr]: %v; oknum: %v; b: %v\n", count, css[fachr], oknum, b)
				if oknum {
					if ok := yield(b, nil); !ok {
						return
					}
				}
			}
		}
	}
}

func CollectBedWithHeader(path string) ([]fastats.BedEntry[[]string], error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	br := bufio.NewReader(r)

	_, err := br.ReadString('\n')

	bed, err := CollectErr[fastats.BedEntry[[]string]](fastats.ParseBed[[]string](br, func(fields []string) ([]string, error) {
		out := make([]string, len(fields))
		copy(out, fields)
		return out, nil
	}))
	return bed, err
}

func FaChrSpanSet(it iter.Seq2[fastats.FaEntry, error]) (map[string]int, error) {
	m := map[string]int{}
	for f, err := range it {
		if err != nil {
			return m, err
		}
		fields := faPosRe.FindStringSubmatch(f.Header)
		m[fields[1]]++
	}

	return m, nil
}

func BedSet[T any](it iter.Seq2[fastats.BedEntry[T], error]) map[string]int {
	m := map[string]int{}
	for f, err := range it {
		if err != nil {
			panic(err)
		}
		m[f.Chr]++
	}
	return m
}

func WriteBed(path string, bed iter.Seq2[fastats.BedEntry[[]string], error]) error {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for f, e := range bed {
		if e != nil {
			return e
		}
		_, e := fmt.Fprintf(bw, "%v\t%v\t%v\t%v\n", f.Chr, f.Start, f.End, strings.Join(f.Fields, "\t"))
		if e != nil {
			return e
		}
	}
	return nil
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

func SliceIter[S ~[]T, T any](s S) iter.Seq[T] {
	return func(y func(T) bool) {
		for _, t := range s {
			if ok := y(t); !ok {
				return
			}
		}
	}
}

func AddErr[T any](it iter.Seq[T]) iter.Seq2[T, error] {
	return func(y func(T, error) bool) {
		for t := range it {
			if ok := y(t, nil); !ok {
				return
			}
		}
	}
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

	pbed := FilterChrParent[[]string](f.Parent, AddErr(SliceIter(bed)))
	bed, e = CollectErr(StripChrParent[[]string](pbed))
	if e != nil {
		panic(e)
	}

	set, e := FaChrSpanSet(AddErr(SliceIter(fa1)))
	if e != nil {
		panic(e)
	}
	SortBed(bed)
	// log.Println("set:", set)

	bedkept, e := CollectErr(KeepBedMatches[[]string](set, AddErr(SliceIter(bed))))
	if e != nil {
		panic(e)
	}

	bedset := BedSet[[]string](AddErr(SliceIter((bedkept))))
	fa1kept, e := CollectErr(KeepFaMatches(bedset, AddErr(SliceIter(fa1))))
	if e != nil {
		panic(e)
	}
	fa2kept, e := CollectErr(KeepFaMatches(bedset, AddErr(SliceIter(fa2))))
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

	if e := WriteBed(f.Outpre + ".bed.gz", AddErr(SliceIter(stripped.Bed))); e != nil {
		panic(e)
	}
	if e := WriteFasta(f.Outpre + "_1.fa.gz", AddErr(SliceIter(stripped.Fa1))); e != nil {
		panic(e)
	}
	if e := WriteFasta(f.Outpre + "_2.fa.gz", AddErr(SliceIter(stripped.Fa2))); e != nil {
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
