package main

import (
	"strconv"
	"fmt"
	"io"
	"os"
	"flag"
	"encoding/csv"
	"bufio"
)

func handle(format string) func(...any) error {
	return func(args ...any) error {
		return fmt.Errorf(format, args...)
	}
}

type Pos struct {
	Chr string
	Start int32
}

func ReadCountBedPath(path string) ([]Pos, map[Pos]int32, error) {
	h := handle("ReadCountBedPath: %w")
	r, e := os.Open(path)
	if e != nil { return nil, nil, h(e) }
	defer r.Close()
	return ReadCountBed(r)
}


func ReadCountBed(r io.Reader) ([]Pos, map[Pos]int32, error) {
	h := handle("ReadCountBed: %w")

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	ps := []Pos{}
	m := map[Pos]int32{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, nil, h(e) }

		if len(line) < 2 {
			return nil, nil, h(fmt.Errorf("len(line) %v < 2", len(line)))
		}

		bp, e := strconv.ParseInt(line[1], 0, 64)
		if e != nil { return nil, nil, h(e) }

		p := Pos{line[0], int32(bp)}

		count, e := strconv.ParseInt(line[3], 0, 64)
		if e != nil { return nil, nil, h(e) }

		m[p] = int32(count)
		ps = append(ps, p)
	}

	return ps, m, nil
}

func GrowCounts(poses []Pos, counts map[Pos]int32, leeway int) ([]Pos, map[Pos]int32) {
	newposes := []Pos{}
	newcounts := map[Pos]int32{}
	for _, pos := range poses {
		count := counts[pos]
		if count > 0 {
			for i := 0; i <= leeway; i++ {
				p := Pos{pos.Chr, pos.Start - int32(i)}
				if _, ok := newcounts[p]; !ok {
					newposes = append(newposes, p)
				}
				// fmt.Println("adding count %v to p %v", count, p)
				newcounts[p] += count

				p = Pos{pos.Chr, pos.Start + int32(i)}
				if _, ok := newcounts[p]; !ok {
					newposes = append(newposes, p)
				}
				// fmt.Println("adding count %v to p %v", count, p)
				newcounts[p] += count
			}
		}
	}
	return newposes, newcounts
}

func FprintHits(w io.Writer, ps []Pos, counter map[Pos]int32) (int, error) {
	n := 0
	for _, pos := range ps {
		posn, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", pos.Chr, pos.Start, pos.Start, counter[pos])
		n += posn
		if e != nil { return n, fmt.Errorf("FprintHits: %w", e) }
	}

	return n, nil
}

func main() {
	leewayp := flag.Int("l", 0, "Allowed distance from cut site")
	flag.Parse()

	poses, counts, e := ReadCountBed(os.Stdin)
	if e != nil { panic(e) }

	gposes, gcounts := GrowCounts(poses, counts, *leewayp)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	_, e = FprintHits(w, gposes, gcounts)
	if e != nil { panic(e) }
}
