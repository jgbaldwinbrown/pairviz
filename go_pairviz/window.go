package main

import (
	"fmt"
	"io"
	"github.com/jgbaldwinbrown/fasttsv"
)

type AllWinStats struct {
	Hits Hits
	GenomeHits GenomeHits
	TotalSelfHits int64
	TotalPairHits int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalReads int64
	Fpkm bool
	Name string
}

type HitSet struct {
	SelfHits int64
	PairHits int64
	SelfFpkm float64
	PairFpkm float64
}

type HitType int

const (
	S HitType = iota
	P
)

type WinHitList []HitSet

func (h WinHitList) IncWin(index int64, hit_type HitType) WinHitList {
	if index < 0 { return h }
	for len(h) <= int(index) {
		h = append(h, HitSet{})
	}
	if hit_type == S {
		h[index].SelfHits++
	} else {
		h[index].PairHits++
	}
	return h
}

type Hits struct {
	Hits map[string]*WinHitList
	WinSize int64
	WinStep int64
}

type GenomeHits struct {
	Ghits map[string]*Hits
	WinSize int64
	WinStep int64
}

type Range struct {
	Start int64
	End int64
	Step int64
}

func (h *Hits) Init(winsize int64, winstep int64) {
	h.Hits = make(map[string]*WinHitList)
	h.WinSize = winsize
	h.WinStep = winstep
}

func (h *Hits) WinsHit(pos int64) (out Range) {
	hiwin := pos / h.WinStep
	out = Range{Start: hiwin, End: hiwin+1, Step: 1}
	for i:=hiwin; ((i * h.WinStep) + h.WinSize) > pos; i-- {
		out.Start = i
	}
	return
}

func (h *Hits) AddHit(chrom string, pos int64, hit_type HitType) {
	_, chromhas := h.Hits[chrom]
	if !chromhas {
		h.Hits[chrom] = new(WinHitList)
	}
	hitwins := h.WinsHit(pos)
	for i:=hitwins.Start; i<hitwins.End; i+=hitwins.Step {
		*h.Hits[chrom] = (*h.Hits[chrom]).IncWin(i, hit_type)
	}
}

func (g *GenomeHits) AddHit(genome string, chrom string, pos int64, hit_type HitType) {
	if _, ok := g.Ghits[genome]; !ok {
		h := new(Hits)
		h.Init(g.WinSize, g.WinStep)
		g.Ghits[genome] = h
	}
	g.Ghits[genome].AddHit(chrom, pos, hit_type)
}

func (g *GenomeHits) Init(winsize, winstep int64) {
	g.Ghits = make(map[string]*Hits)
	g.WinSize = winsize
	g.WinStep = winstep
}

func WinStats(flags Flags, r io.Reader) (stats AllWinStats) {
	stats.Hits.Init(flags.WinSize, flags.WinStep)
	stats.GenomeHits.Init(flags.WinSize, flags.WinStep)
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
		if RangeBad(flags.Distance, flags.MinDistance, flags.PairMinDistance, flags.SelfInMinDistance, pair) {
			continue
		}

		// fmt.Println(pair)
		if pair.Read1.Parent == pair.Read2.Parent {
			stats.Hits.AddHit(pair.Read1.Chrom, pair.Read1.Pos, S)
			stats.Hits.AddHit(pair.Read2.Chrom, pair.Read2.Pos, S)
			stats.GenomeHits.AddHit(pair.Read1.Parent, pair.Read1.Chrom, pair.Read1.Pos, S)
			stats.GenomeHits.AddHit(pair.Read2.Parent, pair.Read2.Chrom, pair.Read2.Pos, S)
		} else {
			stats.Hits.AddHit(pair.Read1.Chrom, pair.Read1.Pos, P)
			stats.Hits.AddHit(pair.Read2.Chrom, pair.Read2.Pos, P)
			stats.GenomeHits.AddHit(pair.Read1.Parent, pair.Read1.Chrom, pair.Read1.Pos, P)
			stats.GenomeHits.AddHit(pair.Read2.Parent, pair.Read2.Chrom, pair.Read2.Pos, P)
		}
		// fmt.Println(stats)
		// for key, val := range stats.Hits.Hits {
		// 	fmt.Println(key, *val)
		// }
	}
	stats.TotalBadReads = stats.TotalReads - stats.TotalGoodReads

	if !flags.NoFpkm {
		stats.Fpkm = true
		for chrom, chromentries := range stats.Hits.Hits {
			for index, win := range *chromentries {
				(*stats.Hits.Hits[chrom])[index].SelfFpkm = Fpkm(win.SelfHits, stats.TotalReads, stats.Hits.WinSize)
				(*stats.Hits.Hits[chrom])[index].PairFpkm = Fpkm(win.PairHits, stats.TotalReads, stats.Hits.WinSize)
			}
		}
	}
	return
}

func FprintWinStats(w io.Writer, stats AllWinStats, separategenomes bool) {
	if separategenomes {
		FprintWinStatsSeparateGenomes(w, stats)
	} else {
		FprintWinStatsPlain(w, stats)
	}
}

func FprintWinStatsPlain(w io.Writer, stats AllWinStats) {
	FprintHeader(w)
	format_string := "%s\t%d\t%d\t%s\t%s\t%d\t%d\t%.8g\t%.8g\t%.8g\t%.8g\t%.8g\t%d\t%d"
	fpkm_format_string := "\t%.8g\t%.8g\t%.8g\t%.8g"
	name_format_string := "\t%s"
	for chrom, chromentries := range stats.Hits.Hits {
		for index, win := range *chromentries {
			fmt.Fprintf(w,
				format_string,
				chrom,
				int64(index) * stats.Hits.WinStep,
				(int64(index) * stats.Hits.WinStep) + stats.Hits.WinSize,
				"paired",
				"self",
				win.PairHits,
				win.SelfHits,
				float64(win.PairHits) / (float64(win.PairHits) + float64(win.SelfHits)),
				float64(win.SelfHits) / (float64(win.PairHits) + float64(win.SelfHits)),
				float64(win.PairHits) / (float64(stats.TotalGoodReads) + float64(stats.TotalBadReads)),
				float64(win.PairHits) / float64(stats.TotalGoodReads),
				float64(win.PairHits) / float64(stats.TotalReads),
				stats.Hits.WinSize,
				stats.Hits.WinStep,
			)
			if stats.Fpkm {
				fmt.Fprintf(w,
					fpkm_format_string,
					win.PairFpkm,
					win.SelfFpkm,
					win.PairFpkm / (win.SelfFpkm + win.PairFpkm),
					win.SelfFpkm / (win.SelfFpkm + win.PairFpkm),
				)
			}
			if stats.Name != "" {
				fmt.Fprintf(w,
					name_format_string,
					stats.Name,
				)
			}
			fmt.Fprintln(w, "")
		}
	}
}

func FprintWinStatsSeparateGenomes(w io.Writer, stats AllWinStats) {
	FprintHeader(w)
	format_string := "%s\t%d\t%d\t%s\t%s\t%d\t%d\t%.8g\t%.8g\t%.8g\t%.8g\t%.8g\t%d\t%d"
	fpkm_format_string := "\t%.8g\t%.8g\t%.8g\t%.8g"
	name_format_string := "\t%s"
	for genome, genomeentries := range stats.GenomeHits.Ghits {
		for chrom, chromentries := range genomeentries.Hits {
			for index, win := range *chromentries {
				fmt.Fprintf(w,
					format_string,
					fmt.Sprintf("%s_%s", chrom, genome),
					int64(index) * genomeentries.WinStep,
					(int64(index) * genomeentries.WinStep) + genomeentries.WinSize,
					"paired",
					"self",
					win.PairHits,
					win.SelfHits,
					float64(win.PairHits) / (float64(win.PairHits) + float64(win.SelfHits)),
					float64(win.SelfHits) / (float64(win.PairHits) + float64(win.SelfHits)),
					float64(win.PairHits) / (float64(stats.TotalGoodReads) + float64(stats.TotalBadReads)),
					float64(win.PairHits) / float64(stats.TotalGoodReads),
					float64(win.PairHits) / float64(stats.TotalReads),
					genomeentries.WinSize,
					genomeentries.WinStep,
				)
				if stats.Fpkm {
					fmt.Fprintf(w,
						fpkm_format_string,
						win.PairFpkm,
						win.SelfFpkm,
						win.PairFpkm / (win.SelfFpkm + win.PairFpkm),
						win.SelfFpkm / (win.SelfFpkm + win.PairFpkm),
					)
				}
				if stats.Name != "" {
					fmt.Fprintf(w,
						name_format_string,
						stats.Name,
					)
				}
				fmt.Fprintln(w, "")
			}
		}
	}
}
