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
}

type ChromStats struct {
	SelfHits map[string]int64
	PairHits map[string]int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalChromosomeReads int64
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
	flag.Parse()

	f.WinSize = int64(wintemp)
	f.WinStep = int64(steptemp)
	f.Distance = int64(disttemp)
	f.GenomeLength = int64(genlentemp)
	f.NameCol = f.Name != ""

	if f.WinSize == -1 || f.WinStep == -1 {
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
	if len(line) < 5 {
		ok = false
		return
	}
	if line[0][0] == '#' {
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

func ChromosomeStats(f Flags, r io.Reader) (stats ChromStats) {
	stats = MakeChromStats()
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		pair, ok := ParsePair(s.Line())
		if !ok {
			continue
		}
		stats.TotalChromosomeReads++
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
		stats.PairHits[pair.Read1.Chrom]++
		stats.TotalGoodReads++
	}
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

func main() {
	flags := GetFlags()
	if flags.Chromosome {
		PrintChromStats(ChromosomeStats(flags, os.Stdin))
	} // else {
		// PrintWinStats(&flags, os.Stdin)
	// }
}
