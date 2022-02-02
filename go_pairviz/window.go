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
	TotalchromosomeReads int64
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

// func WindowedStats(flags Flags, r io.Reader) {
// 	s := fasttsv.NewScanner(r)
// 	for s.Scan() {
// 		
// 	}
		/*
		pair, ok := ParsePair(s.Line())
		if !ok {
			continue
		}
		if f.Distance != -1 && Abs(pair.Read1.Pos - pair.Read2.Pos) > f.Distance {
			stats.TotalGoodReads++
			stats.SelfHits[pair.Read1.Chrom]++
			continue
		}
		if pair.Read1.Chrom != pair.Read2.Chrom || pair.Read1.Parent == pair.Read2.Parent {
			stats.TotalGoodReads++
			stats.SelfHits[pair.Read1.Chrom]++
			continue
		}
		*/
//}
