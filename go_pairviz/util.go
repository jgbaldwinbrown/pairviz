package main

import (
	"os"
	"io"
	"flag"
	"fmt"
	"strings"
	"strconv"
)

type Flags struct {
	WinSize int64
	WinStep int64
	Distance int64
	Chromosome bool
	Name string
	NameCol bool
	Stdin bool
	NoFpkm bool
	Region string
	SeparateGenomes bool
	MinDistance int64
	PairMinDistance int64
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

func Fpkm(count int64, total_sample_reads int64, window_length int64) float64 {
	pmsf := float64(total_sample_reads) / 1e6
	fpm := float64(count) / pmsf
	myfpkm := fpm / (float64(window_length) / 1e3)
	return myfpkm
}

func GetFlags() (f Flags) {
	err := fmt.Errorf("Argument parsing error")
	var wintemp, steptemp, disttemp, mindisttemp, pairmindisttemp int
	flag.StringVar(&f.Name, "n", "", "Name to add to end of table.")
	flag.IntVar(&wintemp, "w", -1, "Window size.")
	flag.IntVar(&steptemp, "s", -1, "Window step distance.")
	flag.IntVar(&disttemp, "d", -1, "Distance between two paired reads before they are ignored.")
	flag.IntVar(&mindisttemp, "m", -1, "Minimum distance between two self reads reads.")
	flag.IntVar(&pairmindisttemp, "pm", -1, "Minimum distance between two paired reads.")
	flag.BoolVar(&f.Stdin, "i", false, "Use Stdin as input (ignored; always do this anyway).")
	flag.BoolVar(&f.Chromosome, "c", false, "Calculate whole-chromosome statistics, not sliding windows.")
	flag.BoolVar(&f.NoFpkm, "f", false, "Do not compute fpkm statistics.")
	flag.StringVar(&f.Region, "r", "", "Calculate statistics in a set of regions specified by this bedfile (not compatible with whole-chromosome statistics or window statistics).")
	flag.BoolVar(&f.SeparateGenomes, "G", false, "Print two entries for each chromosome location, one for each genome, correctly distinguishing self and paired reads (default = false).")

	_ = flag.Int("g", 0, "unused")
	flag.Parse()

	f.WinSize = int64(wintemp)
	f.WinStep = int64(steptemp)
	f.Distance = int64(disttemp)
	f.MinDistance = int64(mindisttemp)
	f.PairMinDistance = int64(pairmindisttemp)
	f.NameCol = f.Name != ""

	if (f.WinSize == -1 || f.WinStep == -1) && !f.Chromosome && f.Region == "" {
		if f.WinSize == -1 {
			fmt.Fprintln(os.Stderr, "Missing -w, winsize")
		}

		if f.WinStep == -1 {
			fmt.Fprintln(os.Stderr, "Missing -s, winstep")
		}

		if !f.Chromosome {
			fmt.Fprintln(os.Stderr, "Missing -c, chromosome analysis")
		}

		if f.Region == "" {
			fmt.Fprintln(os.Stderr, "Missing -r, region")
		}

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

func RangeBad(maxdist int64, mindist int64, pairmindist int64, pair Pair) bool {
	dist := Abs(pair.Read1.Pos - pair.Read2.Pos)

	if pair.Read1.Parent == pair.Read2.Parent {
		return ((mindist != -1 && dist < mindist) ||
			(maxdist != -1 && dist > maxdist) ||
			(pair.Read1.Chrom != pair.Read2.Chrom))
	} else {
		return ((pairmindist != -1 && dist < pairmindist) ||
			(maxdist != -1 && dist > maxdist) ||
			(pair.Read1.Chrom != pair.Read2.Chrom))
	}
}

func FprintHeader(w io.Writer) {
	fmt.Fprintln(w, "chrom\tstart\tend\thit_type\talt_hit_type\thits\talt_hits\tpair_prop\talt_prop\tpair_totprop\tpair_totgoodprop\tpair_totcloseprop\twinsize\twinstep\tpair_fpkm\talt_fpkm\tpair_prop_fpkm\talt_prop_fpkm")
}
