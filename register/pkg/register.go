package register

import (
	"io"
	"strings"
	"log"
	"os"
	"encoding/csv"
	"strconv"
	"fmt"
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
	Dir int
}

type Pair struct {
	Read1 Read
	Read2 Read
}

type Facing int

const (
	Unknown Facing = iota
	In
	Out
	Match
)

func (p Pair) Face() Facing {
	if p.Read1.Dir < 0 && p.Read2.Dir < 0 {
		return Match
	}
	if p.Read1.Dir > 0 && p.Read2.Dir > 0 {
		return Match
	}

	lread := p.Read1
	rread := p.Read2

	if p.Read2.Pos < p.Read1.Pos {
		lread = p.Read2
		rread = p.Read1
	}

	if lread.Dir < 0 && rread.Dir > 0 {
		return Out
	}

	if lread.Dir > 0 && rread.Dir < 0 {
		return In
	}

	return Unknown
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

	switch fields[2] {
		case "+": read.Dir = 1
		case "-": read.Dir = -1
		default: read.Dir = 0
	}

	return
}

func ParsePair(line []string) (pair Pair, ok bool) {
	if !IsAPair(line) {
		ok = false
		return
	}
	pair.Read1 = ParseRead(append([]string{}, line[1], line[2], line[5]))
	pair.Read2 = ParseRead(append([]string{}, line[3], line[4], line[6]))
	return pair, true
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func IsAPair(line []string) bool {
	if len(line) < 7 {
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
	if sl == nil {
		return
	}
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]++
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

	selfincounts := []int64{}
	pairincounts := []int64{}
	transincounts := []int64{}
	selftransincounts := []int64{}
	pairtransincounts := []int64{}

	selfoutcounts := []int64{}
	pairoutcounts := []int64{}
	transoutcounts := []int64{}
	selftransoutcounts := []int64{}
	pairtransoutcounts := []int64{}

	selfmatchcounts := []int64{}
	pairmatchcounts := []int64{}
	transmatchcounts := []int64{}
	selftransmatchcounts := []int64{}
	pairtransmatchcounts := []int64{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		p, ok := ParsePair(line)
		if !ok { log.Printf("pair %v not ok", p); continue }

		if !p.Read1.Ok || !p.Read2.Ok { continue }

		dist := Abs(p.Read2.Pos - p.Read1.Pos)

		face := p.Face()
		var selffacecountp, pairfacecountp, transfacecountp, selftransfacecountp, pairtransfacecountp *[]int64 = nil, nil, nil, nil, nil
		switch face {
		case In:
			selffacecountp = &selfincounts
			pairfacecountp = &pairincounts
			transfacecountp = &transincounts
			selftransfacecountp = &selftransincounts
			pairtransfacecountp = &pairtransincounts
		case Out:
			selffacecountp = &selfoutcounts
			pairfacecountp = &pairoutcounts
			transfacecountp = &transoutcounts
			selftransfacecountp = &selftransoutcounts
			pairtransfacecountp = &pairtransoutcounts
		case Match:
			selffacecountp = &selfmatchcounts
			pairfacecountp = &pairmatchcounts
			transfacecountp = &transmatchcounts
			selftransfacecountp = &selftransmatchcounts
			pairtransfacecountp = &pairtransmatchcounts
		default:
		}

		if p.Read1.Chrom != p.Read2.Chrom {
			SliceInc(&transcounts, dist)
			SliceInc(transfacecountp, dist)
			if p.Read1.Parent == p.Read2.Parent {
				SliceInc(&selftranscounts, dist)
				SliceInc(selftransfacecountp, dist)
			} else {
				SliceInc(&pairtranscounts, dist)
				SliceInc(pairtransfacecountp, dist)
			}
			continue
		}
		if p.Read1.Parent == p.Read2.Parent {
			SliceInc(&selfcounts, dist)
			SliceInc(selffacecountp, dist)
			continue
		}
		SliceInc(&paircounts, dist)
		SliceInc(pairfacecountp, dist)
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
		var si, pi, ti, sti, pti int64 = 0, 0, 0, 0, 0
		var so, po, to, sto, pto int64 = 0, 0, 0, 0, 0
		var sm, pm, tm, stm, ptm int64 = 0, 0, 0, 0, 0

		if len(selfcounts) > i { s = selfcounts[i] }
		if len(paircounts) > i { p = paircounts[i] }
		if len(transcounts) > i { t = transcounts[i] }
		if len(selftranscounts) > i { st = selftranscounts[i] }
		if len(pairtranscounts) > i { pt = pairtranscounts[i] }

		if len(selfincounts) > i { si = selfincounts[i] }
		if len(pairincounts) > i { pi = pairincounts[i] }
		if len(transincounts) > i { ti = transincounts[i] }
		if len(selftransincounts) > i { sti = selftransincounts[i] }
		if len(pairtransincounts) > i { pti = pairtransincounts[i] }

		if len(selfoutcounts) > i { so = selfoutcounts[i] }
		if len(pairoutcounts) > i { po = pairoutcounts[i] }
		if len(transoutcounts) > i { to = transoutcounts[i] }
		if len(selftransoutcounts) > i { sto = selftransoutcounts[i] }
		if len(pairtransoutcounts) > i { pto = pairtransoutcounts[i] }

		if len(selfmatchcounts) > i { sm = selfmatchcounts[i] }
		if len(pairmatchcounts) > i { pm = pairmatchcounts[i] }
		if len(transmatchcounts) > i { tm = transmatchcounts[i] }
		if len(selftransmatchcounts) > i { stm = selftransmatchcounts[i] }
		if len(pairtransmatchcounts) > i { ptm = pairtransmatchcounts[i] }

		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			i,
			p, s, t, st, pt,
			pi, si, ti, sti, pti,
			po, so, to, sto, pto,
			pm, sm, tm, stm, ptm,
		)
	}

	return nil
}
