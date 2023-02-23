package main

import (
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

func ReadCutsiteBedReader(r io.Reader, leeway int) ([]Pos, map[Pos]struct{}, error) {
	h := handle("ReadCutsiteBedReader: %w")
	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	ps := []Pos{}
	m := map[Pos]struct{}{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, nil, h(e) }

		if len(line) < 2 {
			return nil, nil, h(fmt.Errorf("len(line) %v < 2", len(line)))
		}

		bp, e := strconv.ParseInt(line[1], 0, 64)
		if e != nil { return nil, nil, h(e) }

		p := Pos{line[0], bp}
		m[Pos{line[0], bp}] = struct{}{}
		ps = append(ps, p)

		for i := 0; i < leeway; i++ {
			pp := Pos{line[0], bp+int64(i)}
			pm := Pos{line[0], bp-int64(i)}
			m[pp] = struct{}{}
			m[pm] = struct{}{}
		}
	}

	return ps, m, nil
}

func ReadCutsiteBed(path string, leeway int) ([]Pos, map[Pos]struct{}, error) {
	h := handle("ReadCutsiteBed: %w")

	r, e := os.Open(path)
	if e != nil { return nil, nil, h(e) }
	defer r.Close()

	return ReadCutsiteBedReader(r, leeway)
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
		p.Reverse.Chr = line[1]
		p.Reverse.Start, e = strconv.ParseInt(line[4], 0, 64)
		if e != nil { return p, h(e) }
		p.Reverse.Start--
	}

	return p, nil
}

func FilterPairviz(counter map[Pos]struct{}, r io.Reader, w io.Writer) error {
	h := handle("CountHits: %w")

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	cw := csv.NewWriter(w)
	defer cw.Flush()
	cw.Comma = rune('\t')

	commentre := regexp.MustCompile(`^#`)

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) > 0 && commentre.MatchString(line[0]) {
			continue
		}

		p, e := ParsePairsLine(line)
		if e != nil { return h(e) }

		if p.ForwardGood && p.ReverseGood {
			_, okf := counter[p.Forward]
			_, okr := counter[p.Reverse]
			if okr && okf {
				e = cw.Write(line)
				if e != nil { return h(e) }
			}
		}
	}

	return nil
}

func main() {
	cutpathp := flag.String("c", "", "Path to bed file containing cut sites")
	leewayp := flag.Int("l", 0, "Acceptable basepairs from restriction site")
	flag.Parse()

	if *cutpathp == "" {
		panic(fmt.Errorf("missing -c"))
	}

	_, cutsites, e := ReadCutsiteBed(*cutpathp, *leewayp)
	if e != nil { panic(e) }

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	e = FilterPairviz(cutsites, os.Stdin, w)
	if e != nil { panic(e) }
}
