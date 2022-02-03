package main

import (
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
	err := fmt.Errorf("missing argument")
	var wintemp, steptemp, disttemp int
	flag.StringVar(&f.Name, "n", "", "Name to add to end of table.")
	flag.IntVar(&wintemp, "w", -1, "Window size.")
	flag.IntVar(&steptemp, "s", -1, "Window step distance.")
	flag.IntVar(&disttemp, "d", -1, "Distance between two paired reads before they are ignored.")
	flag.BoolVar(&f.Stdin, "i", false, "Use Stdin as input (ignored; always do this anyway).")
	flag.BoolVar(&f.Chromosome, "c", false, "Calculate whole-chromosome statistics, not sliding windows.")
	flag.BoolVar(&f.NoFpkm, "f", false, "Do not compute fpkm statistics.")
	flag.StringVar(&f.Region, "r", "", "Calculate statistics in a set of regions specified by this bedfile (not compatible with whole-chromosome statistics or window statistics).")
	flag.Parse()

	f.WinSize = int64(wintemp)
	f.WinStep = int64(steptemp)
	f.Distance = int64(disttemp)
	f.NameCol = f.Name != ""

	if (f.WinSize == -1 || f.WinStep == -1) && !f.Chromosome && f.Region == "" {
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

func RangeBad(maxdist int64, pair Pair) bool {
	return (maxdist != -1 && Abs(pair.Read1.Pos - pair.Read2.Pos) > maxdist) || (pair.Read1.Chrom != pair.Read2.Chrom)
}
