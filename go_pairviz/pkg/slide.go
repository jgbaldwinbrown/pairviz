package pairviz

import (
	"strconv"
	"io"
	"iter"
	"slices"
	"log"
	"fmt"
	"os"
	"flag"

	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/iterh"
)

type CsvEntry struct {
	Header []string
	Cols map[string]int
	Line []string
}

func CsvIter(r io.Reader) iter.Seq2[CsvEntry, error] {
	return func(yield func(CsvEntry, error) bool) {
		cr := csvh.CsvIn(r)
		cr.ReuseRecord = false
		header, e := cr.Read()
		if e != nil && !yield(CsvEntry{}, e) {
			return
		}
		header = slices.Clone(header)
		cols := csvh.NamesToCols(header)

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			if e != nil && !yield(CsvEntry{}, e) {
				return
			}
			c := CsvEntry{Header: header, Cols: cols, Line: l}
			if !yield(c, nil) {
				return
			}
		}
	}
}

type SlideFlags struct {
	Winsize int
	Winstep int
}

func SumWindow(win iterh.WinView[CsvEntry]) (sum map[string]float64, start, end int) {
	sum = map[string]float64{}
	l := win.Len()
	for i := 0; i < l; i ++ {
		c := win.At(i)
		if len(c.Line) < 2 {
			continue
		}
		if i == 0 {
			var e error
			start, e = strconv.Atoi(c.Line[0])
			if e != nil {
				log.Fatal(e)
			}
		}

		var e error
		end, e = strconv.Atoi(c.Line[0])
		if e != nil {
			log.Fatal(e)
		}
		for j := 1; j < len(c.Line); j++ {
			countstr := c.Line[j]
			count, e := strconv.ParseFloat(countstr, 64)
			if e != nil {
				log.Fatal(e)
			}
			sum[c.Header[j]] += count
		}
	}
	return sum, start, end
}

func WriteHeader(header []string) {
	for i, col := range header {
		if i == 0 {
			if _, e := fmt.Fprintf(os.Stdout, "start\tend"); e != nil {
				log.Fatal(e)
			}
		} else {
			if _, e := fmt.Fprintf(os.Stdout, "\t%v", col); e != nil {
				log.Fatal(e)
			}
		}
	}
	if _, e := fmt.Fprintf(os.Stdout, "\n"); e != nil {
		log.Fatal(e)
	}
}

func WriteWindow(header []string, sum map[string]float64, start, end int) {
	for _, colname := range header {
		if colname == "distance" {
			if _, e := fmt.Fprintf(os.Stdout, "%v\t%v", start, end); e != nil {
				log.Fatal(e)
			}
		} else {
			if _, e := fmt.Fprintf(os.Stdout, "\t%v", sum[colname]); e != nil {
				log.Fatal(e)
			}
		}
	}
	if _, e := fmt.Fprintf(os.Stdout, "\n"); e != nil {
		log.Fatal(e)
	}
}

func HandleWindow(win iterh.WinView[CsvEntry], wini int) {
	sum, start, end := SumWindow(win)

	header := win.At(0).Header
	if wini == 0 {
		WriteHeader(header)
	}
	WriteWindow(header, sum, start, end)
}

func FullSlideLenHist() {
	var f SlideFlags
	flag.IntVar(&f.Winsize, "w", 1, "window size")
	flag.IntVar(&f.Winstep, "s", 1, "window step")
	flag.Parse()
	it, ep := iterh.BreakWithError(CsvIter(os.Stdin))
	wini := 0
	for win := range iterh.Window(it, f.Winsize, f.Winstep) {
		HandleWindow(win, wini)
		wini++
	}
	if *ep != nil {
		log.Fatal(*ep)
	}
}
