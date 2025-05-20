package pairviz

import (
	"io"
	"fmt"
	"math"
	"os"
	"bufio"
	"encoding/csv"
)

func AvgLen(r io.Reader) (count int64, sum float64, mean float64) {
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
			count++
			dist := math.Abs(float64(pair.Read1.Pos - pair.Read2.Pos))
			sum += dist
		}

	}
	mean = sum / float64(count)
	return count, sum, mean
}

func FullAvgLen() {
	count, sum, mean := AvgLen(bufio.NewReader(os.Stdin))
	fmt.Printf("%v\t%v\t%v\n", count, sum, mean)
}
