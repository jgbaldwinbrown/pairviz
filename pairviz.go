package main

import (
	"flag"
	"github.com/jgbaldwinbrown/fasttsv"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"
)

type Flags struct {
	WinSize int64
	WinStep int64
	Distance int64
	Chromosome bool
	GenomeLength int64
	Name string
	NameCol bool
	Stdin bool
	NoFpkm bool
	Region string
}

type ChromStats struct {
	SelfHits map[string]int64
	PairHits map[string]int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalChromosomeReads int64
}

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

type RegionStats struct {
	TotalGoodHits int64
	TotalBadHits int64
	TotalHits int64
	Regions []region
	Fpkm bool
	Name string
}

type Region struct {
	Chrom string
	Start int64
	End int64
	SelfHits int64
	PairHits int64
	SelfFpkm float64
	PairFpkm float64
}

type Read struct {
	Chrom string
	Parent string
	Ok bool
	Pos int64
}

type Pair struct {
	Read1 Read
	Read2 Read
}

type WinHitList []int64

func (h WinHitList) IncWin(index int64) WinHitList {
	while len(h) <= index {
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

type Range struct {
	Start int64
	End int64
	Step int64
}

func Reverse(ints []int64) {
	i := 0
	j := len(ints)-1
	for i < j {
		ints[i], ints[j] = ints[j], ints[i]
		i++
		j--
	}
}

func (h *Hits) WinsHit(pos int64) (out Range) {
	hiwin := pos / winsize
	out = Range{Start: hiwin, End: hiwin+1; Step: 1}
	for i:=hiwin; i * winstep + winsize > pos; i-- {
		out.Start = i
	}
	return
}

func (h *Hits) AddHit(chrom string, pos int64) {
	_, chromhas := h.Hits[chrom]
	if !chromhas {
		h.Hits[chrom] = new(HitList)
	}
	hitwins := h.WinsHit(pos)
	for i:=hitwins.Start; i<hitwins.End; i+=hitwins.Step {
		*h.Hits[chrom] = (*h.Hits[chrom]).IncWin
	}
}

func Fpkm(count int64, total_sample_reads int64, window_length int64) float64 {
	pmsf := float64(total_sample_reads) / 1e6
	fpm := float64(count) / pmsf
	myfpk = fpm / (float64(window_length) / 1e3)
	return myfpkm
}

func GetFlags() (f Flags) {
	err := fmt.Errorf("missing argument")
	var wintemp, steptemp, disttemp, genlentemp int
	flag.StringVar(&f.Name, "n", "", "Name to add to end of table.")
	flag.IntVar(&wintemp, "w", -1, "Window size.")
	flag.IntVar(&steptemp, "s", -1, "Window step distance.")
	flag.IntVar(&disttemp, "d", -1, "Distance between two paired reads before they are ignored.")
	flag.IntVar(&genlentemp, "g", -1, "Genome length (required if using FPKM).")
	flag.BoolVar(&f.Stdin, "i", false, "Use Stdin as input (ignored; always do this anyway).")
	flag.BoolVar(&f.Chromosome, "c", false, "Calculate whole-chromosome statistics, not sliding windows.")
	flag.BoolVar(&f.NoFpkm, "f", false, "Do not compute fpkm statistics.")
	flag.StringVar(&f.Region, "r", "", "Calculate statistics in a set of regions specified by this bedfile (not compatible with whole-chromosome statistics or window statistics).")
	flag.Parse()

	f.WinSize = int64(wintemp)
	f.WinStep = int64(steptemp)
	f.Distance = int64(disttemp)
	f.GenomeLength = int64(genlentemp)
	f.NameCol = f.Name != ""

	if (f.WinSize == -1 || f.WinStep == -1) && !f.Chromosome && f.Region == "" {
		panic(err)
	}
	if f.GenomeLength == -1 && !f.NoFpkm {
		panic(err)
	}
	return
}

func ParseRead(fields []string) (read Read) {
	read.Ok = fields[0] != "!"
	if !read.Ok {
		return
	}
	chrparent := strings.Split(fields[0], "_")
	read.Chrom = chrparent[0]
	read.Parent = chrparent[1]
	var err error
	read.Pos, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return
}

func ParsePair(line []string) (pair Pair, ok bool) {
	if !IsAPair(line) {
		ok = false
		return
	}
	pair.Read1 = ParseRead(line[1:3])
	pair.Read2 = ParseRead(line[3:5])
	return pair, true
}

func MakeChromStats() (stats ChromStats) {
	stats.SelfHits = make(map[string]int64)
	stats.PairHits = make(map[string]int64)
	return
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func IsAPair(line []string) bool {
	if len(line) < 5 {
		return false
	}
	if line[0][0] == '#' {
		return false
	}
	return true
}

func CheckGood(line []string) bool {
	if len(line) < 2 { return false }
	return line[1] != "!"
}

func RangeBad(maxdist int64, pair Pair) bool {
	return (maxdist != -1 && Abs(pair.Read1.Pos - pair.Read2.Pos) > maxdist) || (pair.Read1.Chrom != pair.Read2.Chrom)
}

func ChromosomeStats(f Flags, r io.Reader) (stats ChromStats) {
	stats = MakeChromStats()
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		if IsAPair(line) {
			stats.TotalChromosomeReads++
		}
		if CheckGood(line) {
			stats.TotalGoodReads++
		}

		pair, ok := ParsePair(s.Line())
		if !ok { continue }
		if RangeBad(f.Distance, pair) { continue }

		if pair.Read1.Parent == pair.Read2.Parent {
			if _, inmap := stats.SelfHits[pair.Read1.Chrom]; !inmap {
				stats.SelfHits[pair.Read1.Chrom] = 0
			}
			stats.SelfHits[pair.Read1.Chrom]++
		} else {
			if _, inmap := stats.PairHits[pair.Read1.Chrom]; !inmap {
				stats.PairHits[pair.Read1.Chrom] = 0
			}
			stats.PairHits[pair.Read1.Chrom]++
		}
	}
	stats.TotalBadReads = stats.TotalChromosomeReads - stats.TotalGoodReads
	return
}

func PrintChromStats(stats ChromStats) {
	for k, v := range stats.SelfHits {
		fmt.Printf("Self\t%s\t%d\n", k, v)
	}
	for k, v := range stats.PairHits {
		fmt.Printf("Pair\t%s\t%d\n", k, v)
	}
	for k, v := range stats.PairHits {
		fmt.Printf("Pair propotion:\t%s\t%d\n", k, (v / (v + stats.SelfHits[k])))
	}
	for k, v := range stats.PairHits {
		fmt.Printf("Pair propotion of total good reads:\t%s\t%d\n", k, (v / stats.TotalGoodReads))
	}
	for k, v := range stats.PairHits {
		fmt.Printf("Pair propotion of total reads:\t%s\t%d\n", k, (v / (stats.TotalGoodReads + stats.TotalBadReads)))
	}
}

func FpkmRegion(region *Region, total_chrom_hits int64) {
	winsize := region.End - region.Start
	region.SelfFpkm = Fpkm(region.SelfHits, total_chrom_hits, winsize)
	region.PairFpkm = Fpkm(region.PairHits, total_chrom_hits, winsize)
}

func Overlap(p Pair, r Region) bool {
	if p.Read1.Chrom != p.Read2.Chrom { return false }
	if p.Read1.Chrom != r.Chrom { return false }
	if !((p.Read1.Pos >= r.Start && p.Read1.Pos < r.End) || (p.Read2.Pos >= r.Start && p.Read2.Pos < r.End)) { return false }
	return true
}

func IncrementRegion(p Pair, r *Region) {
	if p.Read1.Parent == p.Read2.Parent {
		r.SelfHits++
	} else {
		r.PairHits++
	}
}

func RegionStats(flags Flags, r io.Reader) (region_stats RegionStats, err error) {
	regions, err = GetRegions(flags.Region)
	if err != nil { return }
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		pair, ok := ParsePair(s.Line())
		if !ok { continue }

		if IsAPair(s.Line()) {
			stats.TotalHits++
		}
		if CheckGood(s.Line()) {
			stats.TotalGoodHits++
		}
		for i, _ := range regions {
			if Overlap(pair, regions[i] && !RangeBad(pair, flags.Distance)) {
				IncrementRegion(pair, regions[i])
			}
		}
	}
	if !flags.NoFpkm {
		for i, _ := range regions {
			FpkmRegion(&regions[i], region_stats.TotalHits)
		}
	}
	stats.
	return
}

func PrintRegionStats(stats RegionStats) {
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
			float64(region.PairHits) / (float64(stats.TotGoodreads) + float64(stats.TotBadreads))
			float64(region.PairHits) / float64(stats.TotGoodreads),
			float64(region.PairHits) / float64(stats.TotChromreads),
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
}

func main() {
	flags := GetFlags()
	if flags.Chromosome {
		PrintChromStats(ChromosomeStats(flags, os.Stdin))
	} //else if flags.Region != "" {
		//regions, err := RegionStats(flags, os.Stdin)
		//if err != nil {panic(err)}
		//PrintRegionStats(regions)
	//} else {
		// PrintWinStats(WinStats(flags, os.Stdin))
	// }
}
