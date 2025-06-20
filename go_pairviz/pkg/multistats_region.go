package pairviz

import (
	"io"
	"iter"
	"encoding/csv"
	"fmt"

	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/span"
	"github.com/jgbaldwinbrown/zfile"
	"github.com/jgbaldwinbrown/iterh"
)

type MultiRegionStats struct {
	Regions []*MultiRegion
	RegionsMap map[string]*span.OrderedSet[*MultiRegion, int64]
	Genos []string
	GenosMap map[string]struct{}
	Chrs []string
	ChrLens map[string]int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalReads int64
}

func MakeMultiRegionStats(rs []*MultiRegion, sets map[string]*span.OrderedSet[*MultiRegion, int64]) MultiRegionStats {
	var m MultiRegionStats
	m.ChrLens = map[string]int64{}
	m.GenosMap = map[string]struct{}{}
	m.Regions = rs
	m.RegionsMap = sets
	return m
}

type MultiRegion struct {
	Chr string
	Span fastats.Span
	Hits map[GenoPair]int64
	LongRangeHits map[GenoPair]int64
}

func (r MultiRegion) Left() int64 {
	return r.Span.Start
}

func (r MultiRegion) Right() int64 {
	return r.Span.End
}

func AddHitMultiRegion(r *MultiRegion, p Pair) {
	r.Hits[OrderedGenoPair(p.Read1.Parent, p.Read2.Parent)]++
}

func AddLongRangeHitMultiRegion(r *MultiRegion, p Pair) {
	r.LongRangeHits[OrderedGenoPair(p.Read1.Parent, p.Read2.Parent)]++
}

func NewRegion(cs fastats.ChrSpan) *MultiRegion {
	return &MultiRegion {
		Chr: cs.Chr,
		Span: fastats.Span{cs.Start, cs.End},
		Hits: map[GenoPair]int64{},
		LongRangeHits: map[GenoPair]int64{},
	}
}

func BuildRegions[CS fastats.ChrSpanner](css iter.Seq[CS]) ([]*MultiRegion, map[string]*span.OrderedSet[*MultiRegion, int64]) {
	allregions := []*MultiRegion{}
	rslices := map[string][]*MultiRegion{}
	rmap := map[string]*span.OrderedSet[*MultiRegion, int64]{}
	for cspanner := range css {
		cs := fastats.ToChrSpan(cspanner)
		r := NewRegion(cs)
		allregions = append(allregions, r)
		rslices[cs.Chr] = append(rslices[cs.Chr], r)
	}
	for chr, regions := range rslices {
		rmap[chr] = span.NewOrderedSet[*MultiRegion, int64](regions)
	}
	return allregions, rmap
}

func BedfileToRegions(path string) (regions []*MultiRegion, set map[string]*span.OrderedSet[*MultiRegion, int64], err error) {
	r, e := zfile.Open(path)
	if e != nil {
		return nil, nil, e
	}
	defer func() {
		e := r.Close()
		if err == nil {
			err = e
		}
	}()
	bed, ep := iterh.BreakWithError(fastats.ParseBedFlat(r))
	regions, set = BuildRegions(bed)
	return regions, set, *ep
}


func Unique[T comparable](ts ...T) []T {
	var out []T
	m := map[T]struct{}{}
	for _, t := range ts {
		if _, ok := m[t]; !ok {
			m[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

func RegionStatsMulti(flags Flags, rs []*MultiRegion, sets map[string]*span.OrderedSet[*MultiRegion, int64], r io.Reader) MultiRegionStats {
	var censorSpans []fastats.ChrSpan
	if flags.CensorPath != "" {
		var e error
		censorSpans, e = ReadChrSpans(flags.CensorPath)
		if e != nil {
			panic(e)
		}
	}

	stats := MakeMultiRegionStats(rs, sets)
	s := csv.NewReader(r)
	s.ReuseRecord = true
	s.LazyQuotes = true
	s.FieldsPerRecord = -1
	s.Comma = '\t'
	for l, e := s.Read(); e != io.EOF; l, e = s.Read() {
		if e != nil {
			panic(e)
		}
		if IsAPair(l) {
			stats.TotalReads++
		}
		if CheckGood(l) {
			stats.TotalGoodReads++
		}

		pair, ok := ParsePair(l)
		if !ok {
			continue
		}
		r1touched := sets[pair.Read1.Chrom].Touched(pair.Read1)
		r2touched := sets[pair.Read2.Chrom].Touched(pair.Read2)
		touched := Unique(append(r1touched, r2touched...)...)

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
			for _, t := range touched {
				AddLongRangeHitMultiRegion(t, pair)
			}
		} else if RangeBad(flags.Distance, flags.MinDistance, flags.PairMinDistance, flags.SelfInMinDistance, pair) {
			continue
		} else {
			for _, t := range touched {
				AddHitMultiRegion(t, pair)
			}
		}
	}
	return stats
}

func WriteRegionsMulti(w io.Writer, m MultiRegionStats) error {
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

	for _, region := range m.Regions {
		if _, e := fmt.Fprintf(w, "%v\t%v\t%v", region.Chr, region.Span.Start, region.Span.End); e != nil {
			return e
		}
		for i, geno1 := range m.Genos {
			for _, geno2 := range m.Genos[i:] {
				gp := OrderedGenoPair(geno1, geno2)
				if _, e := fmt.Fprintf(w, "\t%v", region.Hits[gp]); e != nil {
					return e
				}
				if _, e := fmt.Fprintf(w, "\t%v", region.LongRangeHits[gp]); e != nil {
					return e
				}
			}
		}
		if _, e := fmt.Fprintf(w, "\n"); e != nil {
			return e
		}
	}
	return nil
}
