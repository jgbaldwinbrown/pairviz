package pairviz

import (
	"io"
	"fmt"
	"math"
	"os"
	"bufio"
	"encoding/csv"
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

func FullAvgLen() {
	total, genos := AvgLen(bufio.NewReader(os.Stdin))
	fmt.Printf("%v\t%v\t%v\t%v\n", "total", total.Count, total.Sum, total.Mean())
	for name, g := range genos {
		fmt.Printf("%v\t%v\t%v\t%v\n", name, g.Count, g.Sum, g.Mean())
	}
}
