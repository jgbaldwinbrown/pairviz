package pairviz

import (
	"log"
	"io"
	"fmt"
	"math"
	"os"
	"bufio"
	"encoding/csv"
	"flag"
)

type Counter struct {
	Count int64
	Sum float64
}

func (c Counter) Mean() float64 {
	return c.Sum / float64(c.Count)
}

func AvgLen(r io.Reader) (total Counter, genos map[GenoPair]Counter) {
	genos = map[GenoPair]Counter{}

	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	for l, e := cr.Read(); e == nil; l, e = cr.Read() {
		pair, ok := ParsePair(l)
		if !ok {
			continue
		}

		if pair.Read1.Chrom == pair.Read2.Chrom {
			dist := math.Abs(float64(pair.Read1.Pos - pair.Read2.Pos))

			total.Count++
			total.Sum += dist

			g := OrderedGenoPair(pair.Read1.Parent, pair.Read2.Parent)
			gc := genos[g]
			gc.Count++
			gc.Sum += dist
			genos[g] = gc
		}

	}
	return total, genos
}

type GenoPairDist struct {
	GenoPair
	Dist int64
}

type LenHistStats struct {
	Genos []GenoPair
	Counts map[GenoPairDist]int64
	MaxDist int64
}

func LenHist(r io.Reader, dist int64) (LenHistStats, error) {
	var h LenHistStats
	gm := map[GenoPair]struct{}{}
	h.Counts = map[GenoPairDist]int64{}

	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return h, e
		}
		pair, ok := ParsePair(l)
		if !ok {
			continue
		}

		if pair.Read1.Chrom == pair.Read2.Chrom {
			dist := pair.Read1.Pos - pair.Read2.Pos
			if dist < 0 {
				dist = -dist
			}
			if dist >= h.MaxDist {
				h.MaxDist = dist + 1
			}

			g := OrderedGenoPair(pair.Read1.Parent, pair.Read2.Parent)
			if _, ok := gm[g]; !ok {
				h.Genos = append(h.Genos, g)
				gm[g] = struct{}{}
			}

			h.Counts[GenoPairDist{g, dist}]++
		}

	}
	return h, nil
}

func FullLenHist(r io.Reader, w io.Writer, dist int64) (err error) {
	hist, e := LenHist(os.Stdin, dist)
	if e != nil {
		log.Fatal(e)
	}

	bw := bufio.NewWriter(w)
	defer func() {
		e := bw.Flush()
		if err == nil {
			err = e
		}
	}()

	if _, e := fmt.Fprintf(w, "distance"); e != nil {
		return e
	}
	for _, g := range hist.Genos {
		if _, e := fmt.Fprintf(w, "\t%v", g); e != nil {
			return e
		}
	}
	if _, e := fmt.Fprintf(w, "\n"); e != nil {
		return e
	}
	
	var i int64
	for i = 0; i < hist.MaxDist; i++ {
		if _, e := fmt.Fprintf(w, "%v", i); e != nil {
			return e
		}
		for _, g := range hist.Genos {
			count := hist.Counts[GenoPairDist{g, i}]
			if _, e := fmt.Fprintf(w, "\t%v", count); e != nil {
				return e
			}
		}
		if _, e := fmt.Fprintf(w, "\n"); e != nil {
			return e
		}
	}
	return nil
}

type AvgLenFlags struct {
	Distance int64
}

func FullAvgLen() {
	var f AvgLenFlags
	flag.Int64Var(&f.Distance, "d", -1, "Maximum distance to generate per-bp histogram of distances")
	flag.Parse()

	if f.Distance < 0 {
		total, genos := AvgLen(bufio.NewReader(os.Stdin))
		fmt.Printf("%v\t%v\t%v\t%v\n", "total", total.Count, total.Sum, total.Mean())
		for name, g := range genos {
			fmt.Printf("%v\t%v\t%v\t%v\n", name, g.Count, g.Sum, g.Mean())
		}
	} else {
		if e := FullLenHist(os.Stdin, os.Stdout, f.Distance); e != nil {
			log.Fatal(e)
		}
	}
}
