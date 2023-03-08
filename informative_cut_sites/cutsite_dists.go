package main

import (
	"sort"
	"bufio"
	"regexp"
	"strconv"
	"fmt"
	"io"
	"os"
	"flag"
	"encoding/csv"
)

func handle(format string) func(...any) error {
	return func(args ...any) error {
		return fmt.Errorf(format, args...)
	}
}

type Pos struct {
	Chr string
	Start int64
}

func PosLess(i, j Pos) bool {
	if i.Chr < j.Chr {
		return true
	}
	if j.Chr < i.Chr {
		return false
	}
	return i.Start < j.Start
}

func ReadCutsiteBedReader(r io.Reader) ([]Pos, error) {
	h := handle("ReadCutsiteBedReader: %w")
	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	ps := []Pos{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, h(e) }

		if len(line) < 2 {
			return nil, h(fmt.Errorf("len(line) %v < 2", len(line)))
		}

		bp, e := strconv.ParseInt(line[1], 0, 64)
		if e != nil { return nil, h(e) }

		p := Pos{line[0], bp}
		ps = append(ps, p)
	}

	sort.Slice(ps, func(i, j int) bool {
		return PosLess(ps[i], ps[j])
	})

	return ps, nil
}

func ReadCutsiteBed(path string) ([]Pos, error) {
	h := handle("ReadCutsiteBed: %w")

	r, e := os.Open(path)
	if e != nil { return nil, h(e) }
	defer r.Close()

	return ReadCutsiteBedReader(r)
}

type Pairs struct {
	Forward Pos
	ForwardGood bool
	Reverse Pos
	ReverseGood bool
}

func ParsePairsLine(line []string) (Pairs, error) {
	h := handle("ParsePairsLine: %w")
	var p Pairs
	var e error

	if len(line) != 8 { return p, h(fmt.Errorf("len(line) %v != 8", len(line))) }

	p.ForwardGood = false
	p.ReverseGood = false

	if line[1] != "!" {
		p.ForwardGood = true
		p.Forward.Chr = line[1]
		p.Forward.Start, e = strconv.ParseInt(line[2], 0, 64)
		if e != nil { return p, h(e) }
		p.Forward.Start--
	}

	if line[3] != "!" {
		p.ReverseGood = true
		p.Reverse.Chr = line[3]
		p.Reverse.Start, e = strconv.ParseInt(line[4], 0, 64)
		if e != nil { return p, h(e) }
		p.Reverse.Start--
	}

	return p, nil
}

var pairsCommentre = regexp.MustCompile(`^#`)

type PairsReader struct {
	cr *csv.Reader
	Line []string
	Pairs Pairs
	Err error
}

func NewPairsReader(r io.Reader) *PairsReader {
	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	return &PairsReader{cr: cr}
}

func (r *PairsReader) Read() (Pairs, error) {
	for {
		r.Line, r.Err = r.cr.Read()
		if r.Err != nil {
			r.Pairs = Pairs{}
			return r.Pairs, r.Err
		}

		if len(r.Line) > 0 && pairsCommentre.MatchString(r.Line[0]) {
			continue
		}

		r.Pairs, r.Err = ParsePairsLine(r.Line)
		if r.Err != nil {
			r.Pairs = Pairs{}
			return r.Pairs, r.Err
		}

		return r.Pairs, nil
	}

	return r.Pairs, nil
}

type Dist struct {
	Pos
	Dist int64
}

func FprintDistance(w io.Writer, p Pos, d Dist) (n int, err error) {
	// fmt.Println("Printing p", p, "and d", d)
	n, err = fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", p.Chr, p.Start, p.Start+1, d.Dist, d.Chr, d.Start)
	return n, err
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func FindClosestCutsite(p Pos, cuts []Pos) Dist {
	i := sort.Search(len(cuts), func(i int) bool {
		return PosLess(p, cuts[i])
	})

	if i >= len(cuts) || i <= 0 {
		return Dist{}
	}


	lcut := cuts[i-1]
	rcut := cuts[i]
	if lcut.Chr != p.Chr || rcut.Chr != p.Chr {
		return Dist{}
	}

	ld := Dist{lcut, lcut.Start - p.Start}
	rd := Dist{rcut, rcut.Start - p.Start}
	if Abs(ld.Dist) <= Abs(rd.Dist) {
		return ld
	}
	return rd
}

func CountDistances(sortedCutsites []Pos, r io.Reader, w io.Writer) error {
	h := handle("CountHits: %w")

	pr := NewPairsReader(r)

	for p, e := pr.Read(); e != io.EOF; p, e = pr.Read() {
		if e != nil { return h(e) }

		if p.ForwardGood {
			closest := FindClosestCutsite(p.Forward, sortedCutsites)
			if closest.Chr != "" {
				_, e := FprintDistance(w, p.Forward, closest)
				if e != nil { return h(e) }
			}
		}

		if p.ReverseGood {
			closest := FindClosestCutsite(p.Reverse, sortedCutsites)
			if closest.Chr != "" {
				_, e := FprintDistance(w, p.Reverse, closest)
				if e != nil { return h(e) }
			}
		}
	}

	return nil
}

var stdout = bufio.NewWriter(os.Stdout)

func main() {
	defer stdout.Flush()

	cutpathp := flag.String("c", "", "Path to bed file containing cut sites")
	flag.Parse()

	if *cutpathp == "" {
		panic(fmt.Errorf("missing -c"))
	}

	cutsites, e := ReadCutsiteBed(*cutpathp)
	if e != nil { panic(e) }

	e = CountDistances(cutsites, os.Stdin, stdout)
	if e != nil { panic(e) }
}
