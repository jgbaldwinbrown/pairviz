package main

import (
)

type AllWinStats struct {
	SelfHits Hits
	PairHits Hits
	SelfFpkm Fpkms
	PairFpkms Fpkms
	TotalSelfHits int64
	TotalPairHits int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalReads int64
	Fpkm bool
	Name string
}

type WinHitList []int64
type WinFpkmList []float64

func (h WinHitList) IncWin(index int64) WinHitList {
	for len(h) <= int(index) {
		h = append(h, 0)
	}
	h[index]++
	return h
}

type Hits struct {
	Hits map[string]*WinHitList
	WinSize int64
	WinStep int64
}

type Fpkms struct {
	Fpkms map[string]*WinFpkmList
	WinSize int64
	WinStep int64
}

type Range struct {
	Start int64
	End int64
	Step int64
}

func (h *Hits) WinsHit(pos int64) (out Range) {
	hiwin := pos / h.WinSize
	out = Range{Start: hiwin, End: hiwin+1, Step: 1}
	for i:=hiwin; i * h.WinStep + h.WinSize > pos; i-- {
		out.Start = i
	}
	return
}

func (h *Hits) AddHit(chrom string, pos int64) {
	_, chromhas := h.Hits[chrom]
	if !chromhas {
		h.Hits[chrom] = new(WinHitList)
	}
	hitwins := h.WinsHit(pos)
	for i:=hitwins.Start; i<hitwins.End; i+=hitwins.Step {
		*h.Hits[chrom] = (*h.Hits[chrom]).IncWin(1)
	}
}

func WinStats(flags Flags, r io.Reader) (stats AllWinStats) {
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
		if RangeBad(f.Distance, pair) {
			continue
		}

		if pair.Read1.Chrom == pair.Read2.Chrom {
			stats.SelfHits.AddHit(pair.Read1.Chrom, pair.Read1.Pos)
		} else {
			stats.PairHits.AddHit(pair.Read1.Chrom, pair.Read1.Pos)
		}
	}
	stats.TotalBadReads = stats.TotalReads - stats.TotalGoodReads

	if !flags.NoFpkm {
	}
}

func PrintWinStats(stats AllWinStats) {
	format_string := "%s\t%d\t%d\t%s\t%s\t%d\t%d\t%.8g\t%.8g\t%.8g\t%.8g\t%.8g\t%d\t%d"
	fpkm_format_string := "\t%.8g\t%.8g\t%.8g\t%.8g"
	name_format_string := "\t%s"
	for _, region := range stats.Regions {
		fmt.Printf(
			format_string,
			region.Chrom,
			region.Start,
			region.End,
			"paired",
			"self",
			region.PairHits,
			region.SelfHits,
			float64(region.PairHits) / (float64(region.PairHits) + float64(region.SelfHits)),
			float64(region.SelfHits) / (float64(region.PairHits) + float64(region.SelfHits)),
			float64(region.PairHits) / (float64(stats.TotalGoodHits) + float64(stats.TotalBadHits)),
			float64(region.PairHits) / float64(stats.TotalGoodHits),
			float64(region.PairHits) / float64(stats.TotalHits),
			region.End - region.Start,
			region.End - region.Start,
		)
		if stats.Fpkm {
			fmt.Printf(
				fpkm_format_string,
				region.PairFpkm,
				region.SelfFpkm,
				region.PairFpkm / (region.SelfFpkm + region.PairFpkm),
				region.SelfFpkm / (region.SelfFpkm + region.PairFpkm),
			)
		}
		if stats.Name != "" {
			fmt.Printf(
				name_format_string,
				stats.Name,
			)
		}
		fmt.Println("")
	}
}
