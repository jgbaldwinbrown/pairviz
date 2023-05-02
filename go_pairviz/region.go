package main

import (
	"strconv"
	"github.com/jgbaldwinbrown/fasttsv"
	"fmt"
	"io"
	"os"
)

type RegionStats struct {
	TotalGoodHits int64
	TotalBadHits int64
	TotalHits int64
	Regions []Region
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

func ParseRegion(line []string) (region Region, err error) {
	region_err := fmt.Errorf("Region parse error")
	if len(line) < 3 {
		err = region_err
		return
	}
	region.Chrom = line[0]
	region.Start, err = strconv.ParseInt(line[1], 0, 64)
	if err != nil { return }
	region.End, err = strconv.ParseInt(line[2], 0, 64)
	return
}

func GetRegions(path string) (regions []Region, err error) {
	r, err := os.Open(path)
	if err != nil { return }
	defer r.Close()

	s := fasttsv.NewScanner(r)
	for s.Scan() {
		var region Region
		region, err = ParseRegion(s.Line())
		if err != nil { return }
		regions = append(regions, region)
	}

	return
}

func GetRegionStats(flags Flags, r io.Reader) (stats RegionStats, err error) {
	stats.Regions, err = GetRegions(flags.Region)
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
		for i, _ := range stats.Regions {
			if Overlap(pair, stats.Regions[i]) && !RangeBad(flags.Distance, flags.MinDistance, flags.PairMinDistance, flags.SelfInMinDistance, pair) {
				IncrementRegion(pair, &stats.Regions[i])
			}
		}
	}
	if !flags.NoFpkm {
		stats.Fpkm = true
		for i, _ := range stats.Regions {
			FpkmRegion(&stats.Regions[i], stats.TotalHits)
		}
	}
	stats.TotalBadHits = stats.TotalHits - stats.TotalGoodHits
	return
}

func FprintRegionStats(w io.Writer, stats RegionStats) {
	FprintHeader(w, true, -1, false)
	format_string := "%s\t%d\t%d\t%s\t%s\t%d\t%d\t%.8g\t%.8g\t%.8g\t%.8g\t%.8g\t%d\t%d"
	fpkm_format_string := "\t%.8g\t%.8g\t%.8g\t%.8g"
	name_format_string := "\t%s"
	for _, region := range stats.Regions {
		fmt.Fprintf(w,
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
			fmt.Fprintf(w,
				fpkm_format_string,
				region.PairFpkm,
				region.SelfFpkm,
				region.PairFpkm / (region.SelfFpkm + region.PairFpkm),
				region.SelfFpkm / (region.SelfFpkm + region.PairFpkm),
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
