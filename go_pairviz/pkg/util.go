package pairviz

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
	SelfInMinDistance int64
	ReadLen int64
	JsonOut bool
}

type Read struct {
	Chrom string
	Parent string
	Ok bool
	Pos int64
	Dir int
}

type Facing int

const (
	Unknown Facing = iota
	In
	Out
	Match
)

type Pair struct {
	Read1 Read
	Read2 Read
}

func (p Pair) AbsPosDist() int64 {
	return Abs(p.Read2.Pos - p.Read1.Pos)
}

func (p Pair) Face() Facing {
	if p.Read1.Dir < 0 && p.Read2.Dir < 0 {
		return Match
	}
	if p.Read1.Dir > 0 && p.Read2.Dir > 0 {
		return Match
	}

	lread := p.Read1
	rread := p.Read2

	if p.Read2.Pos < p.Read1.Pos {
		lread = p.Read2
		rread = p.Read1
	}

	if lread.Dir < 0 && rread.Dir > 0 {
		return Out
	}

	if lread.Dir > 0 && rread.Dir < 0 {
		return In
	}

	return Unknown
}

func Fpkm(count int64, total_sample_reads int64, window_length int64) float64 {
	pmsf := float64(total_sample_reads) / 1e6
	fpm := float64(count) / pmsf
	myfpkm := fpm / (float64(window_length) / 1e3)
	// log.Printf("Fpkm: count: %v; total_sample_reads: %v; window_length: %v pmsf: %v; fpm: %v; myfpkm: %v\n", count, total_sample_reads, window_length, pmsf, fpm, myfpkm)
	return myfpkm
}

func GetFlags() (f Flags) {
	err := fmt.Errorf("Argument parsing error")
	var wintemp, steptemp, disttemp, mindisttemp, pairmindisttemp, selfinmindisttemp, readlentemp int
	flag.StringVar(&f.Name, "n", "", "Name to add to end of table.")
	flag.IntVar(&wintemp, "w", -1, "Window size.")
	flag.IntVar(&steptemp, "s", -1, "Window step distance.")
	flag.IntVar(&disttemp, "d", -1, "Distance between two paired reads before they are ignored.")
	flag.IntVar(&mindisttemp, "m", -1, "Minimum distance between two self reads reads.")
	flag.IntVar(&pairmindisttemp, "pm", -1, "Minimum distance between two paired reads.")
	flag.IntVar(&selfinmindisttemp, "sim", -1, "Minimum distance between inward-facing self reads.")
	flag.BoolVar(&f.Stdin, "i", false, "Use Stdin as input (ignored; always do this anyway).")
	flag.BoolVar(&f.Chromosome, "c", false, "Calculate whole-chromosome statistics, not sliding windows.")
	flag.BoolVar(&f.NoFpkm, "f", false, "Do not compute fpkm statistics.")
	flag.StringVar(&f.Region, "r", "", "Calculate statistics in a set of regions specified by this bedfile (not compatible with whole-chromosome statistics or window statistics).")
	flag.BoolVar(&f.SeparateGenomes, "G", false, "Print two entries for each chromosome location, one for each genome, correctly distinguishing self and paired reads (default = false).")
	flag.IntVar(&readlentemp, "rlen", -1, "Length of reads in pairs (used to calculate overlapping or not; skipped otherwise).")
	flag.BoolVar(&f.JsonOut, "j", false, "Output as JSON")

	_ = flag.Int("g", 0, "unused")
	flag.Parse()

	f.WinSize = int64(wintemp)
	f.WinStep = int64(steptemp)
	f.Distance = int64(disttemp)
	f.MinDistance = int64(mindisttemp)
	f.PairMinDistance = int64(pairmindisttemp)
	f.SelfInMinDistance = int64(selfinmindisttemp)
	f.ReadLen = int64(readlentemp)
	f.NameCol = f.Name != ""
	fmt.Fprintf(os.Stderr, "flag Name: %v; NameCol: %v\n", f.Name, f.NameCol)

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
	read.Parent = "ecoli"
	if len(chrparent) >= 2 {
		read.Parent = chrparent[1]
	}
	var err error
	read.Pos, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}

	switch fields[2] {
		case "+": read.Dir = 1
		case "-": read.Dir = -1
		default: read.Dir = 0
	}

	return
}

func ParsePair(line []string) (pair Pair, ok bool) {
	if !IsAPair(line) {
		ok = false
		return
	}
	pair.Read1 = ParseRead(append([]string{}, line[1], line[2], line[5]))
	pair.Read2 = ParseRead(append([]string{}, line[3], line[4], line[6]))
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

func (p Pair) IsSelfIn() bool {
	return p.Read1.Chrom == p.Read2.Chrom && p.Read1.Parent == p.Read2.Parent && p.Face() == In
}

func RangeBad(maxdist int64, mindist int64, pairmindist int64, selfinmindist int64, pair Pair) bool {
	dist := Abs(pair.Read1.Pos - pair.Read2.Pos)
	if pair.IsSelfIn() && dist < selfinmindist {
		return true
	}

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

func FprintHeader(w io.Writer, fpkm bool, readlen int64, namecol bool) {
	fmt.Fprintf(os.Stderr, "Header namecol: %v\n", namecol)
	fmt.Fprint(w, "chrom\tstart\tend\thit_type\talt_hit_type\thits\talt_hits\tpair_prop\talt_prop\tpair_totprop\tpair_totgoodprop\tpair_totcloseprop\twinsize\twinstep")
	if fpkm {
		fmt.Fprint(w, "\tpair_fpkm\talt_fpkm\tpair_prop_fpkm\talt_prop_fpkm")
	}
	if readlen != -1 {
		fmt.Fprint(w, "\tovl\tnon_ovl\tovl_prop\tnon_ovl_prop")
		if fpkm {
			fmt.Fprint(w, "\tovl_fpkm\tnon_ovl_fpkm\tovl_prop_fpkm\tnon_ovl_prop_fpkm")
		}
	}
	if namecol {
		fmt.Fprint(w, "\tname")
	}
	fmt.Fprintln(w, "")
}
