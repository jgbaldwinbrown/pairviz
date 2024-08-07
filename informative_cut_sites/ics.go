package main

import (
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

func ReadCutsiteBed(path string) ([]Pos, map[Pos]int64, error) {
	h := handle("ReadCutsiteBed: %w")

	r, e := os.Open(path)
	if e != nil { return nil, nil, h(e) }
	defer r.Close()

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	ps := []Pos{}
	m := map[Pos]int64{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, nil, h(e) }

		if len(line) < 2 {
			return nil, nil, h(fmt.Errorf("len(line) %v < 2", len(line)))
		}

		bp, e := strconv.ParseInt(line[1], 0, 64)
		if e != nil { return nil, nil, h(e) }

		p := Pos{line[0], bp}
		m[Pos{line[0], bp}] = 0
		ps = append(ps, p)
	}

	return ps, m, nil
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

	if line[1] == "!" {
		p.ForwardGood = false
	} else {
		p.ForwardGood = true
		p.Forward.Chr = line[1]
		p.Forward.Start, e = strconv.ParseInt(line[2], 0, 64)
		if e != nil { return p, h(e) }
		p.Forward.Start--
	}

	if line[3] == "!" {
		p.ReverseGood = false
	} else {
		p.ReverseGood = true
		p.Reverse.Chr = line[3]
		p.Reverse.Start, e = strconv.ParseInt(line[4], 0, 64)
		if e != nil { return p, h(e) }
		p.Reverse.Start--
	}

	return p, nil
}

func CountHitsOld(counter map[Pos]int64, r io.Reader) error {
	h := handle("CountHits: %w")

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	commentre := regexp.MustCompile(`^#`)

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) > 0 && commentre.MatchString(line[0]) {
			continue
		}

		p, e := ParsePairsLine(line)
		if e != nil { return h(e) }

		if p.ForwardGood {
			if _, ok := counter[p.Forward]; ok {
				counter[p.Forward]++
			}
		}

		if p.ReverseGood {
			if _, ok := counter[p.Reverse]; ok {
				counter[p.Reverse]++
			}
		}
	}

	return nil
}

func CountHits(counter map[Pos]int64, r io.Reader, leeway int) error {
	h := handle("CountHits: %w")

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	commentre := regexp.MustCompile(`^#`)

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) > 0 && commentre.MatchString(line[0]) {
			continue
		}

		p, e := ParsePairsLine(line)
		if e != nil { return h(e) }

		if p.ForwardGood {
			if _, ok := counter[p.Forward]; ok {
				for i := 0; i <= leeway; i++ {
					fpos := Pos{p.Forward.Chr, p.Forward.Start + int64(i)}
					counter[fpos]++
					fpos = Pos{p.Forward.Chr, p.Forward.Start - int64(i)}
					counter[fpos]++
				}
			}
		}

		if p.ReverseGood {
			if _, ok := counter[p.Reverse]; ok {
				for i := 0; i <= leeway; i++ {
					rpos := Pos{p.Reverse.Chr, p.Reverse.Start + int64(i)}
					counter[rpos]++
					rpos = Pos{p.Reverse.Chr, p.Reverse.Start - int64(i)}
					counter[rpos]++
				}
			}
		}
	}

	return nil
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

func FprintHits(w io.Writer, ps []Pos, counter map[Pos]int64) (int, error) {
	n := 0
	for _, pos := range ps {
		posn, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", pos.Chr, pos.Start, pos.Start, counter[pos])
		n += posn
		if e != nil { return n, fmt.Errorf("FprintHits: %w", e) }
	}

	return n, nil
}

func main() {
	cutpathp := flag.String("c", "", "Path to bed file containing cut sites")
	leewayp := flag.Int("l", 0, "Allowed distance from cut site")
	flag.Parse()

	if *cutpathp == "" {
		panic(fmt.Errorf("missing -c"))
	}

	poses, cutsites, e := ReadCutsiteBed(*cutpathp)
	if e != nil { panic(e) }

	e = CountHits(cutsites, os.Stdin, *leewayp)
	if e != nil { panic(e) }

	// fmt.Fprintln(os.Stderr, cutsites)

	_, e = FprintHits(os.Stdout, poses, cutsites)
	if e != nil { panic(e) }
}
