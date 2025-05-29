package pairviz

import (
	"flag"
	"io"
	"strconv"
	"fmt"
	"os"
	"log"

	"github.com/jgbaldwinbrown/csvh"
)

func LongRangeRatioPerSite(r io.Reader, w io.Writer, numIdCols int) (err error) {
	cw := csvh.CsvOut(w)
	defer func() {
		cw.Flush()
		csvh.DeferE(&err, cw.Error())
	}()
	var out []string
	cri := 0
	for ent, e := range CsvIter(r) {
		if e != nil {
			return e
		}
		if cri == 0 {
			out = append(out[:0], ent.Header[:numIdCols]...)
			for i := numIdCols; i < len(ent.Header); i += 2 {
				out = append(out, ent.Header[i])
			}
			if e := cw.Write(out); e != nil {
				return e
			}
		}
		out = append(out[:0], ent.Line[:numIdCols]...)
		for i := numIdCols; i < len(ent.Line); i += 2 {
			sr, e := strconv.ParseFloat(ent.Line[i], 64)
			if e != nil {
				return e
			}
			lr, e := strconv.ParseFloat(ent.Line[i+1], 64)
			if e != nil {
				return e
			}
			out = append(out, fmt.Sprintf("%v", sr / lr))
		}
		if e := cw.Write(out); e != nil {
			return e
		}

		cri++
	}
	return nil
}

type longRangeRatioFlags struct {
	NumIdCols int
}

func FullLongRangeRatio() {
	var f longRangeRatioFlags
	flag.IntVar(&f.NumIdCols, "n", 3, "number of ID columns (columns to ignore)")
	flag.Parse()
	if e := LongRangeRatioPerSite(os.Stdin, os.Stdout, f.NumIdCols); e != nil {
		log.Fatal(e)
	}
}
