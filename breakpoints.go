package main

import (
	"fmt"
	"os"
	"github.com/jgbaldwinbrown/fasttsv"
	"io"
	"bufio"
	"strconv"
	"strings"
)

type BedEntry struct {
	Chrom string
	Start int64
	End int64
	Fields []string
}

func ParseBedLine(l []string) (b BedEntry, e error) {
	if len(l) < 3 {
		e = fmt.Errorf("Bed line too short")
		return
	}
	b.Chrom = l[0]
	b.Start, e = strconv.ParseInt(l[1], 0, 64)
	if e != nil { return }
	b.End, e = strconv.ParseInt(l[2], 0, 64)
	if e != nil { return }
	if len(l) > 3 {
		b.Fields = append([]string{}, l[3:]...)
	}
	return
}

func Breakpoints(b BedEntry, dist int64) (bp1, bp2 BedEntry) {
	bp1 = b
	bp2 = b

	bp1.Start = b.Start - dist
	bp1.End = b.Start + dist + 1
	bp1.Fields = append([]string{"left_breakpoint"}, b.Fields...)

	bp2.Start = b.End - dist - 1
	bp2.End = b.End + dist
	bp2.Fields = append([]string{"right_breakpoint"}, b.Fields...)

	return
}

func (b BedEntry) Fprintln(w io.Writer) {
	fmt.Fprintf(w, "%v\t%v\t%v", b.Chrom, b.Start, b.End)
	if b.Fields != nil && len(b.Fields) > 0 {
		fields := strings.Join(b.Fields, "\t")
		fmt.Fprintf(w, "\t%v", fields)
	}
	fmt.Fprintln(w, "")
}

func BedBreakpoints(r io.Reader, w io.Writer) {
	s := fasttsv.NewScanner(r)
	for s.Scan() {
		bentry, err := ParseBedLine(s.Line())
		if err != nil { continue }

		bp1, bp2 := Breakpoints(bentry, 5000)
		bp1.Fprintln(w)
		bp2.Fprintln(w)
	}
}

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	BedBreakpoints(os.Stdin, w)
}
