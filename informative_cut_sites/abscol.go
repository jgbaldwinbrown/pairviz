package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"fmt"
	"io"
)

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	cr := csv.NewReader(os.Stdin)
	cr.Comma = rune('\t')
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.ReuseRecord = true

	cw := csv.NewWriter(os.Stdout)
	cw.Comma = rune('\t')
	defer cw.Flush()

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { panic(e) }
		if len(l) > 3 {
			i, e := strconv.ParseInt(l[3], 0, 64)
			if e == nil {
				l[3] = fmt.Sprint(Abs(i))
			}
		}
		e = cw.Write(l)
		if e != nil { panic(e) }
	}
}
