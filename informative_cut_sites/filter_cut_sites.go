package main

import (
	"bufio"
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

func ReadCutsiteBedReader(r io.Reader) ([]Pos, map[Pos]struct{}, error) {
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
		m[p] = struct{}{}
		ps = append(ps, p)
	}

	return ps, m, nil
}

func ReadCutsiteBed(path string) ([]Pos, map[Pos]struct{}, error) {
	h := handle("ReadCutsiteBed: %w")

	r, e := os.Open(path)
	if e != nil { return nil, nil, h(e) }
	defer r.Close()

	return ReadCutsiteBedReader(r)
}

func FilterCutsites(cutsites map[Pos]struct{}, r io.Reader, w io.Writer, thresh int) error {
	h := handle("ReadCounts: %w")
	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	cw := csv.NewWriter(w)
	defer cw.Flush()
	cw.Comma = rune('\t')

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		// fmt.Fprintf(stderr, "line: %v\n", line)
		if e != nil { return h(e) }

		if len(line) < 4 {
			return h(fmt.Errorf("len(line) %v < 4", len(line)))
		}

		bp, e := strconv.ParseInt(line[1], 0, 64)
		if e != nil { return h(e) }

		p := Pos{line[0], bp}

		count, e := strconv.ParseInt(line[3], 0, 64)
		if e != nil { return h(e) }

		// fmt.Fprintf(stderr, "p: %v; count: %v\n", p, count)
		_, ok := cutsites[p]
		if ok && count >= int64(thresh) {
			// fmt.Fprintf(stderr, "writing line: %v\n", line)
			cw.Write(line)
		}

	}

	return nil
}

var stderr *bufio.Writer

func main() {
	stderr = bufio.NewWriter(os.Stderr)
	defer stderr.Flush()

	cutpathp := flag.String("c", "", "Path to bed file containing cut sites")
	threshp := flag.Int("t", 0, "Threshold of coverage to keep")
	flag.Parse()

	if *cutpathp == "" {
		panic(fmt.Errorf("missing -c"))
	}

	_, cutsites, e := ReadCutsiteBed(*cutpathp)
	if e != nil { panic(e) }
	// fmt.Fprintf(stderr, "cutsites: %v\n", cutsites)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	e = FilterCutsites(cutsites, os.Stdin, w, *threshp)
	if e != nil { panic(e) }
}
