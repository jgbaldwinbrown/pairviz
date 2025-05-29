package pairviz

import (
	"flag"
	"os"
	"bufio"
	"strconv"
	"io"
	"iter"
	"regexp"
	"fmt"
	"log"

	"github.com/jgbaldwinbrown/fastats/pkg"
)

func Format(x, y string) string {
	if x <= y {
		return x + "_" + y
	}
	return y + "_" + x
}

var gpBracketRe = regexp.MustCompile(`{([^ ]*) ([^}]*)}`)
var gpRe = regexp.MustCompile(`([^_]+)_(.+)`)

func ParseGenoPair(raw string) (GenoPair, error) {
	fields := gpRe.FindStringSubmatch(raw)
	if len(fields) < 3 {
		return GenoPair{}, fmt.Errorf("ParseGenoPair: could not parse %v", raw)
	}
	return GenoPair{fields[1], fields[2]}, nil
}

func LongRangeRatio(counts map[string]int64, focal, sameSpecies, otherSpecies1, otherSpecies2 string) float64 {
	totalLong := (
		float64(counts[Format(focal, sameSpecies) + "_longrange"]) * 2.0 + 
		float64(counts[Format(focal, otherSpecies1) + "_longrange"]) + 
		float64(counts[Format(focal, otherSpecies1) + "_longrange"]))
	expect := totalLong * 100000.0 / 130000000.0
	actual := (
		float64(counts[Format(focal, sameSpecies)]) * 2.0 + 
		float64(counts[Format(focal, otherSpecies1)]) + 
		float64(counts[Format(focal, otherSpecies1)]))
	// log.Printf("totalLong %v; expect %v; actual %v\n", totalLong, expect, actual)
	return actual / expect
}

func SpeciesRatio(counts map[string]int64, focal, sameSpecies, otherSpecies1, otherSpecies2 string) float64 {
	within := float64(counts[Format(focal, sameSpecies)])
	between1 := float64(counts[Format(focal, otherSpecies1)])
	between2 := float64(counts[Format(focal, otherSpecies2)])
	// log.Printf("within %v; between1 %v; between2 %v\n", within, between1, between2)
	return within / ((between1 + between2) / 2.0)
}

type GenoCount struct {
	Geno string
	Count float64
}

type GenoPairCount struct {
	GenoPair
	Count float64
}

func GenoCounts(pairs ...GenoPairCount) map[string]float64 {
	m := map[string]float64{}
	for _, p := range pairs {
		m[p.Geno1] += p.Count
		m[p.Geno2] += p.Count
	}
	return m
}

func GenoPairExpectedFreqs(genos ...GenoCount) map[GenoPair]float64 {
	total := 0.0
	for _, gc := range genos {
		total += gc.Count
	}
	
	m := map[GenoPair]float64{}
	for i, gc1 := range genos {
		for _, gc2 := range genos[i:] {
			gp := GenoPair{gc1.Geno, gc2.Geno}
			m[gp] = (gc1.Count / total) * (gc2.Count / total)
		}
	}
	return m
}

func GenoPairExpectedCounts(pairs ...GenoPairCount) []GenoPairCount {
	genos := GenoCounts(pairs...)
	genos2 := make([]GenoCount, 0, len(genos))
	for key, val := range genos {
		genos2 = append(genos2, GenoCount{Geno: key, Count: val})
	}
	
	freqs := GenoPairExpectedFreqs(genos2...)
	total := 0.0
	for _, pair := range pairs {
		total += pair.Count
	}
	counts := make([]GenoPairCount, 0, len(freqs))
	for _, p := range pairs {
		gp := p.GenoPair
		f := freqs[gp]
		counts = append(counts, GenoPairCount {
			GenoPair: gp,
			Count: f * total,
		})
	}
	return counts
}

func GenoPairChisq(pairs ...GenoPairCount) (chisq, p float64) {
	expect := GenoPairExpectedCounts(pairs...)
	oe := make([]ObservedExpected, 0, len(expect))
	for i, p := range pairs {
		oe = append(oe, ObservedExpected{Observed: p.Count, Expected: expect[i].Count})
	}
	chisq = ChiSq(oe...)
	return chisq, ChiSqP(chisq, float64(len(pairs) - 1))
}

func GenoPairChisqNoDivZero(pairs ...GenoPairCount) (expect []GenoPairCount, chisq, p float64, skipped int) {
	expect = GenoPairExpectedCounts(pairs...)
	oe := make([]ObservedExpected, 0, len(expect))
	for i, p := range pairs {
		oe = append(oe, ObservedExpected{Observed: p.Count, Expected: expect[i].Count})
	}
	chisq, skipped = ChiSqNoDivZero(oe...)
	return expect, chisq, ChiSqP(chisq, float64(len(pairs) - 1 - skipped)), skipped
}

type ChiP struct {
	ChiSq float64
	P float64
	Expect []GenoPairCount
	Actual []GenoPairCount
}

func ReadGenoPairChisq(r io.Reader) iter.Seq2[fastats.BedEntry[ChiP], error] {
	return func(yield func(fastats.BedEntry[ChiP], error) bool) {
		ci := CsvIter(r)
		var be fastats.BedEntry[ChiP]
		for ent, e := range ci {
			if e != nil && !yield(be, e) {
				return
			}
			gpc := make([]GenoPairCount, 0, len(ent.Line) - 3)
			for i := 3; i < len(ent.Line); i++ {
				col := ent.Line[i]
				count, e := strconv.ParseFloat(col, 64)
				if e != nil && !yield(be, e) {
					return
				}
				gp, e := ParseGenoPair(ent.Header[i])
				if e != nil && !yield(be, e) {
					return
				}
				gpc = append(gpc, GenoPairCount{GenoPair: OrderedGenoPair(gp.Geno1, gp.Geno2), Count: count})
			}
			expect, chi, p, _ := GenoPairChisqNoDivZero(gpc...)
			start, e := strconv.ParseInt(ent.Line[1], 0, 64)
			if e != nil {
				return
			}
			end, e := strconv.ParseInt(ent.Line[2], 0, 64)
			if e != nil {
				return
			}
			b := fastats.BedEntry[ChiP]{
				ChrSpan: fastats.ChrSpan{Chr: ent.Line[0], Span: fastats.Span{start, end}},
				Fields: ChiP{chi, p, expect, gpc},
			}
			if !yield(b, nil) {
				return
			}
		}
	}
}

type GenoPairChisqFlags struct {
	PrintExpect bool
}

func FullGenoPairChisq() {
	var f GenoPairChisqFlags
	flag.BoolVar(&f.PrintExpect, "e", false, "print expected and actual counts, and ratio, for each genotype")
	flag.Parse()
	
	bw := bufio.NewWriter(os.Stdout)
	defer func() {
		if e := bw.Flush(); e != nil {
			log.Fatal(e)
		}
	}()
	if _, e := fmt.Fprintf(bw, "chr\tstart\tend\tchisq\tp"); e != nil {
		log.Fatal(e)
	}

	bedi := 0
	bed := ReadGenoPairChisq(os.Stdin)
	for b, e := range bed {
		if bedi == 0 {
			if f.PrintExpect {
				for _, expect := range b.Fields.Expect {
					if _, e := fmt.Fprintf(bw, "\t%v_%v", expect.Geno1, expect.Geno2); e != nil {
						log.Fatal(e)
					}
				}
			}
			if _, e := fmt.Fprintf(bw, "\n"); e != nil {
				log.Fatal(e)
			}
		}
		if e != nil {
			log.Fatal(e)
		}
		if _, e := fmt.Fprintf(bw, "%v\t%v\t%v\t%v\t%v", b.Chr, b.Start, b.End, b.Fields.ChiSq, b.Fields.P); e != nil {
			log.Fatal(e)
		}
		if f.PrintExpect {
			for i, expect := range b.Fields.Expect {
				actual := b.Fields.Actual[i]
				if _, e := fmt.Fprintf(bw, "\t%v,%v,%v", expect.Count, actual.Count, actual.Count / expect.Count); e != nil {
					log.Fatal(e)
				}
			}
		}
		if _, e := fmt.Fprintf(bw, "\n"); e != nil {
			log.Fatal(e)
		}

		bedi++
	}
}
