package main

import (
	"io"
	"strings"
	"log"
	"os"
	"encoding/csv"
	"strconv"
	"fmt"
	"bufio"
)

func handle(format string) func(...any) error {
	return func(args ...any) error {
		return fmt.Errorf(format, args...)
	}
}

type Read struct {
	Chrom string
	Parent string
	Ok bool
	Pos int64
}

type Pair struct {
	Read1 Read
	Read2 Read
}

func ParseRead(fields []string) (read Read) {
	read.Ok = fields[0] != "!"
	if !read.Ok {
		return
	}
	chrparent := strings.Split(fields[0], "_")
	read.Chrom = chrparent[0]
	read.Parent = chrparent[1]
	var err error
	read.Pos, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return
}

func ParsePair(line []string) (pair Pair, ok bool) {
	if !IsAPair(line) {
		ok = false
		return
	}
	pair.Read1 = ParseRead(line[1:3])
	pair.Read2 = ParseRead(line[3:5])
	return pair, true
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func IsAPair(line []string) bool {
	if len(line) < 5 {
		return false
	}
	if line[0][0] == '#' {
		return false
	}
	return true
}

func CheckGood(line []string) bool {
	if len(line) < 2 { return false }
	return line[1] != "!"
}

var stdout *os.File

func GrowSlice[T any](sl []T, newlen int64) []T {
	if len(sl) < int(newlen) {
		nsl := make([]T, newlen, newlen*2)
		copy(nsl, sl)
		return nsl
	}
	return sl
}

func SlicePut[T any](sl *[]T, idx int64, val T) {
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]=val
}

func SliceInc(sl *[]int64, idx int64) {
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]++
}

func main() {
	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	e := Run(os.Stdin, stdout)
	if e != nil { panic(e) }
}

func Run(r io.Reader, w io.Writer) error {
	h := handle("Run: %w")
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.Comma = rune('\t')
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	selfcounts := []int64{}
	paircounts := []int64{}
	transcounts := []int64{}
	selftranscounts := []int64{}
	pairtranscounts := []int64{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		p, ok := ParsePair(line)
		if !ok { log.Printf("pair %v not ok", p); continue }

		if !p.Read1.Ok || !p.Read2.Ok { continue }

		dist := Abs(p.Read2.Pos - p.Read1.Pos)

		if p.Read1.Chrom != p.Read2.Chrom {
			SliceInc(&transcounts, dist)
			if p.Read1.Parent == p.Read2.Parent {
				SliceInc(&selftranscounts, dist)
			} else {
				SliceInc(&pairtranscounts, dist)
			}
			continue
		}
		if p.Read1.Parent == p.Read2.Parent {
			SliceInc(&selfcounts, dist)
			continue
		}
		SliceInc(&paircounts, dist)
	}

	maxlen := len(transcounts)
	if len(paircounts) > maxlen {
		maxlen = len(paircounts)
	}
	if len(selfcounts) > maxlen {
		maxlen = len(selfcounts)
	}
	if len(selftranscounts) > maxlen {
		maxlen = len(selftranscounts)
	}
	if len(pairtranscounts) > maxlen {
		maxlen = len(pairtranscounts)
	}

	for i := 0; i < maxlen; i++ {
		var s, p, t, st, pt int64 = 0, 0, 0, 0, 0
		if len(selfcounts) > i { s = selfcounts[i] }
		if len(paircounts) > i { p = paircounts[i] }
		if len(transcounts) > i { t = transcounts[i] }
		if len(selftranscounts) > i { st = selftranscounts[i] }
		if len(pairtranscounts) > i { pt = pairtranscounts[i] }
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", i, p, s, t, st, pt)
	}

	return nil
}
