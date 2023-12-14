package prepfa

import (
	"log"
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
				log.Printf("ok\n")
				oknum := count < css[b.Chr]
				counts[b.Chr]++
				log.Printf("count: %v; css[b.Chr]: %v; oknum: %v; b: %v\n", count, css[b.Chr], oknum, b)
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

func RunCleanup() {
	fa1p := flag.String("fa1", "", "path to .fa or .fa.gz file")
	fa2p := flag.String("fa2", "", "path to .fa or .fa.gz file")
	bedp := flag.String("bed", "", "path to .bed or .bed.gz file")
	outprep := flag.String("o", "out", "output prefix")
	parentp := flag.String("p", "", "parent to keep")
	flag.Parse()

	if *fa1p == "" {
		panic(fmt.Errorf("mising -fa1"))
	}
	if *fa2p == "" {
		panic(fmt.Errorf("mising -fa2"))
	}
	if *bedp == "" {
		panic(fmt.Errorf("missing -bed"))
	}
	if *parentp == "" {
		panic(fmt.Errorf("missing -p"))
	}

	fa1, e := CollectFa(*fa1p)
	if e != nil {
		panic(e)
	}
	SortFa(fa1)

	fa2, e := CollectFa(*fa2p)
	if e != nil {
		panic(e)
	}
	if e != nil {
		panic(e)
	}
	SortFa(fa2)

	bed, e := CollectBedWithHeader(*bedp)
	if e != nil {
		panic(e)
	}

	pbed := FilterChrParent[[]string](*parentp, iter.SliceIter[fastats.BedEntry[[]string]](bed))
	bed, e = iter.Collect[fastats.BedEntry[[]string]](StripChrParent[[]string](pbed))
	if e != nil {
		panic(e)
	}

	set, e := FaChrSpanSet(iter.SliceIter[fastats.FaEntry](fa1))
	if e != nil {
		panic(e)
	}
	SortBed(bed)
	log.Println("set:", set)

	bedkept := KeepBedMatches[[]string](set, iter.SliceIter[fastats.BedEntry[[]string]](bed))
	if e := WriteBed(*outprep + ".bed.gz", bedkept); e != nil {
		panic(e)
	}

	if e := WriteFasta(*outprep + "_1.fa.gz", iter.SliceIter[fastats.FaEntry](fa1)); e != nil {
		panic(e)
	}
	if e := WriteFasta(*outprep + "_2.fa.gz", iter.SliceIter[fastats.FaEntry](fa2)); e != nil {
		panic(e)
	}

}
