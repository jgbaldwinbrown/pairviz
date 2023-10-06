package pairviz

import (
	"github.com/jgbaldwinbrown/fasttsv"
	"fmt"
	"io"
)

type ChromStats struct {
	SelfHits map[string]int64
	PairHits map[string]int64
	TotalBadReads int64
	TotalGoodReads int64
	TotalChromosomeReads int64
}

func MakeChromStats() (stats ChromStats) {
	stats.SelfHits = make(map[string]int64)
	stats.PairHits = make(map[string]int64)
	return
}

func ChromosomeStats(f Flags, r io.Reader) (stats ChromStats) {
	stats = MakeChromStats()
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		if IsAPair(s.Line()) {
			stats.TotalChromosomeReads++
		}
		if CheckGood(s.Line()) {
			stats.TotalGoodReads++
		}

		pair, ok := ParsePair(s.Line())
		if !ok { continue }
		if RangeBad(f.Distance, f.MinDistance, f.PairMinDistance, f.SelfInMinDistance, pair) { continue }

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

func FprintChromStats(w io.Writer, stats ChromStats) {
	for k, v := range stats.SelfHits {
		fmt.Fprintf(w, "Self\t%s\t%d\n", k, v)
	}
	for k, v := range stats.PairHits {
		fmt.Fprintf(w, "Pair\t%s\t%d\n", k, v)
	}
	for k, v := range stats.PairHits {
		fmt.Fprintf(w, "Pair propotion:\t%s\t%d\n", k, (v / (v + stats.SelfHits[k])))
	}
	for k, v := range stats.PairHits {
		fmt.Fprintf(w, "Pair propotion of total good reads:\t%s\t%d\n", k, (v / stats.TotalGoodReads))
	}
	for k, v := range stats.PairHits {
		fmt.Fprintf(w, "Pair propotion of total reads:\t%s\t%d\n", k, (v / (stats.TotalGoodReads + stats.TotalBadReads)))
	}
}

