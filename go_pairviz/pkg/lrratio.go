package pairviz

import (
	"io"
	"strconv"
	"fmt"
	"os"
	"log"

	"github.com/jgbaldwinbrown/csvh"
)

func LongRangeRatioPerSite(r io.Reader, w io.Writer) (err error) {
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
			out = append(out[:0], "chr", "start", "end")
			for i := 3; i < len(ent.Header); i += 2 {
				out = append(out, ent.Header[i])
			}
			if e := cw.Write(out); e != nil {
				return e
			}
		}
		out = append(out[:0], ent.Line[:3]...)
		for i := 3; i < len(ent.Line); i += 2 {
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

func FullLongRangeRatio() {
	if e := LongRangeRatioPerSite(os.Stdin, os.Stdout); e != nil {
		log.Fatal(e)
	}
}
