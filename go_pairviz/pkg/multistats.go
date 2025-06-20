package pairviz

import (
	"github.com/jgbaldwinbrown/fastats/pkg"
	"io"
	"github.com/jgbaldwinbrown/fasttsv"
	"fmt"
	"github.com/jgbaldwinbrown/iterh"
)

type GenoPair struct {
	Geno1 string
	Geno2 string
}

type GenoChrSpan struct {
	fastats.ChrSpan
	GenoPair
}

type MultiWinStats struct {
	Hits map[GenoChrSpan]int64
	LongRangeHits map[GenoChrSpan]int64
	Genos []string
	GenosMap map[string]struct{}
	Chrs []string
	ChrLens map[string]int64
	Winsize int64
	Winstep int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalReads int64
}

func MakeMultiWinStats() MultiWinStats {
	var m MultiWinStats
	m.Hits = map[GenoChrSpan]int64{}
	m.LongRangeHits = map[GenoChrSpan]int64{}
	m.GenosMap = map[string]struct{}{}
	m.ChrLens = map[string]int64{}
	return m
}

func  WinsHitMulti(pos, winsize, winstep int64) (out Range) {
	hiwin := pos / winstep
	out = Range{Start: hiwin, End: hiwin+1, Step: 1}
	for i:=hiwin; ((i * winstep) + winsize) > pos; i-- {
		out.Start = i
	}
	return
}

func OrderedGenoPair(s1, s2 string) GenoPair {
	if s1 <= s2 {
		return GenoPair{s1, s2}
	}
	return GenoPair{s2, s1}
}

func RangeTooLong(maxdist int64, pair Pair) bool {
	if pair.Read1.Chrom != pair.Read2.Chrom {
		return true
	}
	dist := Abs(pair.Read1.Pos - pair.Read2.Pos)
	if dist > maxdist {
		return true
	}
	return false
}


func AddHitMultiCore(hits map[GenoChrSpan]int64, winsize, winstep int64, p Pair) {
	gp := OrderedGenoPair(p.Read1.Parent, p.Read2.Parent)

	hitwins := WinsHitMulti(p.Read1.Pos, winsize, winstep)
	for i := hitwins.Start; i < hitwins.End; i += hitwins.Step {
		start := i * winstep
		gcp := GenoChrSpan {
			GenoPair: gp,
			ChrSpan: fastats.ChrSpan {
				Chr: p.Read1.Chrom,
				Span: fastats.Span {
					Start: start,
					End: start + winsize,
				},
			},
		}
		hits[gcp]++
	}

	hitwins = WinsHitMulti(p.Read2.Pos, winsize, winstep)
	for i := hitwins.Start; i < hitwins.End; i += hitwins.Step {
		start := i * winstep
		gcp := GenoChrSpan {
			GenoPair: gp,
			ChrSpan: fastats.ChrSpan {
				Chr: p.Read2.Chrom,
				Span: fastats.Span {
					Start: start,
					End: start + winsize,
				},
			},
		}
		hits[gcp]++
	}
}

func AddHitMulti(m *MultiWinStats, p Pair) {
	AddHitMultiCore(m.Hits, m.Winsize, m.Winstep, p)
}

func AddLongRangeHitMulti(m *MultiWinStats, p Pair) {
	AddHitMultiCore(m.LongRangeHits, m.Winsize, m.Winstep, p)
}

func OverlapChrSpan(r Read, c fastats.ChrSpan) bool {
	if r.Chrom != c.Chr {
		return false
	}
	if r.Pos < c.Start {
		return false
	}
	return r.Pos < c.End
}

func ReadWithinSpans(OkSpans []fastats.ChrSpan, r Read) bool {
	for _, span := range OkSpans {
		if OverlapChrSpan(r, span) {
			return true
		}
	}
	return false
}

func PairWithinSpans(OkSpans []fastats.ChrSpan, p Pair) bool {
	return ReadWithinSpans(OkSpans, p.Read1) && ReadWithinSpans(OkSpans, p.Read2)
}

func ReadChrSpans(path string) ([]fastats.ChrSpan, error) {
	biter := iterh.PathIter(path, fastats.ParseBedFlat)
	b, e := iterh.CollectWithError(biter)
	if e != nil {
		return nil, e
	}
	cs := make([]fastats.ChrSpan, 0, len(b))
	for _, ent := range b {
		cs = append(cs, ent.ChrSpan)
	}
	return cs, nil
}

func WinStatsMulti(flags Flags, r io.Reader) MultiWinStats {
	var censorSpans []fastats.ChrSpan
	if flags.CensorPath != "" {
		var e error
		censorSpans, e = ReadChrSpans(flags.CensorPath)
		if e != nil {
			panic(e)
		}
	}

	stats := MakeMultiWinStats()
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		if IsAPair(s.Line()) {
			stats.TotalReads++
		}
		if CheckGood(s.Line()) {
			stats.TotalGoodReads++
		}

		pair, ok := ParsePair(s.Line())
		if !ok {
			continue
		}

		if _, ok := stats.GenosMap[pair.Read1.Parent]; !ok {
			stats.GenosMap[pair.Read1.Parent] = struct{}{}
			stats.Genos = append(stats.Genos, pair.Read1.Parent)
		}
		if _, ok := stats.GenosMap[pair.Read2.Parent]; !ok {
			stats.GenosMap[pair.Read2.Parent] = struct{}{}
			stats.Genos = append(stats.Genos, pair.Read2.Parent)
		}

		chrlen, ok := stats.ChrLens[pair.Read1.Chrom]
		if !ok {
			chrlen = pair.Read1.Pos + 1
			stats.ChrLens[pair.Read1.Chrom] = chrlen
			stats.Chrs = append(stats.Chrs, pair.Read1.Chrom)
		}
		if pair.Read1.Pos >= chrlen {
			stats.ChrLens[pair.Read1.Chrom] = pair.Read1.Pos + 1
		}
		chrlen, ok = stats.ChrLens[pair.Read2.Chrom]
		if !ok {
			chrlen = pair.Read2.Pos + 1
			stats.ChrLens[pair.Read2.Chrom] = chrlen
			stats.Chrs = append(stats.Chrs, pair.Read2.Chrom)
		}
		if pair.Read2.Pos >= chrlen {
			stats.ChrLens[pair.Read2.Chrom] = pair.Read2.Pos + 1
		}

		if len(censorSpans) > 0 && !PairWithinSpans(censorSpans, pair) {
			continue
		}

		if RangeTooLong(flags.Distance, pair) {
			AddLongRangeHitMulti(&stats, pair)
		} else if RangeBad(flags.Distance, flags.MinDistance, flags.PairMinDistance, flags.SelfInMinDistance, pair) {
			continue
		} else {
			AddHitMulti(&stats, pair)
		}
	}
	return stats
}

func WriteWindowsMulti(w io.Writer, m MultiWinStats) error {
	if _, e := fmt.Fprintf(w, "chr\tstart\tend"); e != nil {
		return e
	}
	for i, geno1 := range m.Genos {
		for _, geno2 := range m.Genos[i:] {
			gp := OrderedGenoPair(geno1, geno2)
			if _, e := fmt.Fprintf(w, "\t%v_%v\t%v_%v_longrange", gp.Geno1, gp.Geno2, gp.Geno1, gp.Geno2); e != nil {
				return e
			}
		}
	}
	if _, e := fmt.Fprintf(w, "\n"); e != nil {
		return e
	}

	for _, chr := range m.Chrs {
		chrlen := m.ChrLens[chr]
		var start int64
		for start = 0; start + m.Winsize < chrlen; start += m.Winstep {
			end := start + m.Winsize
			if _, e := fmt.Fprintf(w, "%v\t%v\t%v", chr, start, end); e != nil {
				return e
			}
			for i, geno1 := range m.Genos {
				for _, geno2 := range m.Genos[i:] {
					gcp := GenoChrSpan {
						GenoPair: OrderedGenoPair(geno1, geno2),
						ChrSpan: fastats.ChrSpan {Chr: chr, Span: fastats.Span{Start: start, End: end}},
					}
					if _, e := fmt.Fprintf(w, "\t%v", m.Hits[gcp]); e != nil {
						return e
					}
					if _, e := fmt.Fprintf(w, "\t%v", m.LongRangeHits[gcp]); e != nil {
						return e
					}
				}
			}
			if _, e := fmt.Fprintf(w, "\n"); e != nil {
				return e
			}
		}
	}
	return nil
}

