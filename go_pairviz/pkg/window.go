package pairviz

import (
	"bytes"
	"math"
	"os"
	"encoding/json"
	"fmt"
	"io"
	"github.com/jgbaldwinbrown/fasttsv"
)

// All statistics associated with one window
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

// The counts of hits in a single window
type HitSet struct {
	SelfHits int64
	PairHits int64
	OvlHits int64
	NonOvlHits int64
	SelfFpkm float64
	PairFpkm float64
	OvlFpkm float64
	NonOvlFpkm float64
}

// Types are self, pair, overlapped, and non-overlapped
type HitType int

const (
	S HitType = iota
	P
	Ovl
	NonOvl
)

// The raw slice of all hits in all windows along a chromosome
type WinHitList []HitSet

// Increment the correct hit type for the specified index of the WinHitList
func (h WinHitList) IncWin(index int64, hit_type HitType) WinHitList {
	if index < 0 { return h }
	for len(h) <= int(index) {
		h = append(h, HitSet{})
	}
	switch hit_type {
	case S:
		h[index].SelfHits++
	case P:
		h[index].PairHits++
	case Ovl:
		h[index].OvlHits++
	case NonOvl:
		h[index].NonOvlHits++
	}
	return h
}

// The structure containing hit counts for all windows in all chromosomes in
// the genome, plus the window size and step information needed to decode the
// hits.
type Hits struct {
	Hits map[string]*WinHitList
	WinSize int64
	WinStep int64
}

// A wrapper for a set of Hits to allow multiple genomes to be independently
// specified and to match chromosomes between them.
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
	// if hit_type == Ovl || hit_type == NonOvl {
	// 	log.Printf("Hits AddHit: chrom %v; pos %v; hit_type %v\n", chrom, pos, hit_type)
	// }
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
	// if hit_type == Ovl || hit_type == NonOvl {
	// 	log.Printf("GenomeHits AddHit: genome %v; chrom %v; pos %v; hit_type %v\n", genome, chrom, pos, hit_type)
	// }
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
	stats.Name = flags.Name
	stats.Hits.Init(flags.WinSize, flags.WinStep)
	stats.GenomeHits.Init(flags.WinSize, flags.WinStep)
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		if IsAPair(s.Line()) {
			stats.TotalReads++
			// log.Printf("Current total reads: %v", stats.TotalReads)
		}
		if CheckGood(s.Line()) {
			stats.TotalGoodReads++
			// log.Printf("Current total good reads: %v", stats.TotalGoodReads)
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

		if flags.ReadLen != -1 {
			// log.Println("checking overlaps")
			if PairOverlaps(pair, flags.ReadLen) {
				// log.Println("overlapped")
				stats.Hits.AddHit(pair.Read1.Chrom, pair.Read1.Pos, Ovl)
				stats.Hits.AddHit(pair.Read2.Chrom, pair.Read2.Pos, Ovl)
				stats.GenomeHits.AddHit(pair.Read1.Parent, pair.Read1.Chrom, pair.Read1.Pos, Ovl)
				stats.GenomeHits.AddHit(pair.Read2.Parent, pair.Read2.Chrom, pair.Read2.Pos, Ovl)
			} else {
				// log.Println("nonoverlapped")
				stats.Hits.AddHit(pair.Read1.Chrom, pair.Read1.Pos, NonOvl)
				stats.Hits.AddHit(pair.Read2.Chrom, pair.Read2.Pos, NonOvl)
				stats.GenomeHits.AddHit(pair.Read1.Parent, pair.Read1.Chrom, pair.Read1.Pos, NonOvl)
				stats.GenomeHits.AddHit(pair.Read2.Parent, pair.Read2.Chrom, pair.Read2.Pos, NonOvl)
			}
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
				(*stats.Hits.Hits[chrom])[index].OvlFpkm = Fpkm(win.OvlHits, stats.TotalReads, stats.Hits.WinSize)
				(*stats.Hits.Hits[chrom])[index].NonOvlFpkm = Fpkm(win.NonOvlHits, stats.TotalReads, stats.Hits.WinSize)
			}
		}

		for genome, genomeentries := range stats.GenomeHits.Ghits {
			for chrom, chromentries := range genomeentries.Hits {
				for index, win := range *chromentries {
					(*stats.GenomeHits.Ghits[genome].Hits[chrom])[index].SelfFpkm = Fpkm(win.SelfHits, stats.TotalReads, stats.Hits.WinSize)
					(*stats.GenomeHits.Ghits[genome].Hits[chrom])[index].PairFpkm = Fpkm(win.PairHits, stats.TotalReads, stats.Hits.WinSize)
					(*stats.GenomeHits.Ghits[genome].Hits[chrom])[index].OvlFpkm = Fpkm(win.OvlHits, stats.TotalReads, stats.Hits.WinSize)
					(*stats.GenomeHits.Ghits[genome].Hits[chrom])[index].NonOvlFpkm = Fpkm(win.NonOvlHits, stats.TotalReads, stats.Hits.WinSize)
				}
			}
		}
	}
	return
}

// A special variant of float64 that can marshal and unmarshal NaN and Inf values
type JsonFloat float64

func (j JsonFloat) MarshalJSON() ([]byte, error) {
	f := float64(j)
	if math.IsNaN(f) {
		return json.Marshal("NaN")
	}
	if math.IsInf(f, 1) {
		return json.Marshal("Inf")
	}
	if math.IsInf(f, -1) {
		return json.Marshal("-Inf")
	}
	return json.Marshal(f)
}

var jsonNaN = []byte(`"NaN"`)
var jsonInf = []byte(`"Inf"`)
var jsonNegInf = []byte(`"-Inf"`)

func (j *JsonFloat) UnmarshalJSON(p []byte) error {
	if bytes.Compare(p, jsonNaN) == 0 {
		*j = JsonFloat(math.NaN())
		return nil
	}
	if bytes.Compare(p, jsonInf) == 0 {
		*j = JsonFloat(math.Inf(1))
		return nil
	}
	if bytes.Compare(p, jsonNegInf) == 0 {
		*j = JsonFloat(math.Inf(-1))
		return nil
	}
	err := json.Unmarshal(p, (*float64)(j))
	if err != nil {
		return fmt.Errorf("JsonFloat UnmarshalJSON: input %v: %w", string(p), err)
	}
	return nil
}

// An alternate version of the output statistics struct that can convert to JSON without crashing on NaN and Inf values
type JsonOutStat struct {
	Genome string
	Chr string
	Start int64
	End int64
	TargetType string
	AltType string
	TargetHits JsonFloat
	AltHits JsonFloat
	TargetProp JsonFloat
	AltProp JsonFloat
	TargetPropGoodBad JsonFloat
	TargetPropGood JsonFloat
	TargetPropTotal JsonFloat
	WinSize int64
	WinStep int64
	TargetFpkm JsonFloat
	AltFpkm JsonFloat
	TargetFpkmProp JsonFloat
	AltFpkmProp JsonFloat
	AltOvlHits JsonFloat
	AltNonOvlHits JsonFloat
	AltOvlProp JsonFloat
	AltNonOvlProp JsonFloat
	AltOvlFpkm JsonFloat
	AltNonOvlFpkm JsonFloat
	AltOvlFpkmProp JsonFloat
	AltNonOvlFpkmProp JsonFloat
	Name string
}

// Generate the statistics for one window for JSON output
func MakeJsonOutStat(genome, chr, name string, stats AllWinStats, winsize, winstep, index int64, win HitSet) JsonOutStat {
	var j JsonOutStat

	j.Genome = genome
	j.Chr = chr

	j.Start = int64(index) * winstep
	j.End = (int64(index) * winstep) + winsize
	j.TargetType = "paired"
	j.AltType = "self"
	j.TargetHits = JsonFloat(float64(win.PairHits))
	j.AltHits = JsonFloat(float64(win.SelfHits))
	j.TargetProp = JsonFloat(float64(win.PairHits) / (float64(win.PairHits) + float64(win.SelfHits)))
	j.AltProp = JsonFloat(float64(win.SelfHits) / (float64(win.PairHits) + float64(win.SelfHits)))
	j.TargetPropGoodBad = JsonFloat(float64(win.PairHits) / (float64(stats.TotalGoodReads) + float64(stats.TotalBadReads)))
	j.TargetPropGood = JsonFloat(float64(win.PairHits) / float64(stats.TotalGoodReads))
	j.TargetPropTotal = JsonFloat(float64(win.PairHits) / float64(stats.TotalReads))
	j.WinSize = winsize
	j.WinStep = winstep

	j.TargetFpkm = JsonFloat(win.PairFpkm)
	j.AltFpkm = JsonFloat(win.SelfFpkm)
	j.TargetFpkmProp = JsonFloat(win.PairFpkm / (win.SelfFpkm + win.PairFpkm))
	j.AltFpkmProp = JsonFloat(win.SelfFpkm / (win.SelfFpkm + win.PairFpkm))

	j.AltOvlHits = JsonFloat(float64(win.OvlHits))
	j.AltNonOvlHits = JsonFloat(float64(win.NonOvlHits))
	j.AltOvlProp = JsonFloat(float64(win.OvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)))
	j.AltNonOvlProp = JsonFloat(float64(win.NonOvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)))

	j.AltOvlFpkm = JsonFloat(win.OvlFpkm)
	j.AltNonOvlFpkm = JsonFloat(win.NonOvlFpkm)
	j.AltOvlFpkmProp = JsonFloat(float64(win.OvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)))
	j.AltNonOvlFpkmProp = JsonFloat(float64(win.NonOvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)))

	j.Name = name

	return j
}

// Panic on error
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Write all stats for all windows as JSON
func FprintWinStatsJson(w io.Writer, stats AllWinStats, readlen int64) {
	enc := json.NewEncoder(w)

	for genome, genomeentries := range stats.GenomeHits.Ghits {
		for chrom, chromentries := range genomeentries.Hits {
			for index, win := range *chromentries {
				j := MakeJsonOutStat(genome, chrom, stats.Name, stats, genomeentries.WinSize, genomeentries.WinStep, int64(index), win)
				err := enc.Encode(j)
				Must(err)
			}
		}
	}
}

// Write all stats for all windows (wrapper of other Fprint functions)
func FprintWinStats(w io.Writer, stats AllWinStats, separategenomes bool, readlen int64, jsonOut bool) {
	if jsonOut {
		FprintWinStatsJson(w, stats, readlen)
	} else if separategenomes {
		FprintWinStatsSeparateGenomes(w, stats, readlen)
	} else {
		FprintWinStatsPlain(w, stats, readlen)
	}
}

// Write all stats as tab-separated text
func FprintWinStatsPlain(w io.Writer, stats AllWinStats, readlen int64) {
	fmt.Fprintf(os.Stderr, "WinStatsPlain name: %v\n", stats.Name)
	FprintHeader(w, stats.Fpkm, readlen, stats.Name != "")
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

			if readlen != -1 {
				fmt.Fprintf(w,
					"\t%v\t%v\t%v\t%v",
					win.OvlHits,
					win.NonOvlHits,
					float64(win.OvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)),
					float64(win.NonOvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)),
				)
				if stats.Fpkm {
					fmt.Fprintf(w,
						"\t%v\t%v\t%v\t%v",
						win.OvlFpkm,
						win.NonOvlFpkm,
						float64(win.OvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)),
						float64(win.NonOvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)),
					)
				}
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

// Write all stats as tab-separated text, and write stats separately for each genome
func FprintWinStatsSeparateGenomes(w io.Writer, stats AllWinStats, readlen int64) {
	fmt.Fprintf(os.Stderr, "WinStatsSeparateGenomes name: %v\n", stats.Name)
	FprintHeader(w, stats.Fpkm, readlen, stats.Name != "")
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

				if readlen != -1 {
					fmt.Fprintf(w,
						"\t%v\t%v\t%v\t%v",
						win.OvlHits,
						win.NonOvlHits,
						float64(win.OvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)),
						float64(win.NonOvlHits) / (float64(win.OvlHits) + float64(win.NonOvlHits)),
					)
					if stats.Fpkm {
						fmt.Fprintf(w,
							"\t%v\t%v\t%v\t%v",
							win.OvlFpkm,
							win.NonOvlFpkm,
							float64(win.OvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)),
							float64(win.NonOvlFpkm) / (float64(win.OvlFpkm) + float64(win.NonOvlFpkm)),
						)
					}
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
