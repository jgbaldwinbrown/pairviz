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

// The statistics for one read
type Read struct {
	Chrom string
	Parent string
	Ok bool
	Pos int64
	Dir int
}

// The statistics for one read pair
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

// Read types can be self, pair, or trans-chromosome
func (p Pair) ReadType() ReadType {
	if p.Read1.Chrom != p.Read2.Chrom {
		return TransType
	}
	if p.Read1.Parent == p.Read2.Parent {
		return SelfType
	}
	return PairType
}

// Parse a .pairs file read into a Read; panic on error
func ParseRead(fields []string) (read Read) {
	read.Ok = fields[0] != "!"
	if !read.Ok {
		return
	}
	chrparent := strings.Split(fields[0], "_")
	read.Chrom = chrparent[0]
	read.Parent = "ecoli"
	if len(chrparent) >= 2 {
		read.Parent = chrparent[1]
	}
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

// Parse both reads for a line into a read pair; panic on error
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

// Grow the length of a slice to newlen, and the capacity to newlen * 2
func GrowSlice[T any](sl []T, newlen int64) []T {
	if len(sl) < int(newlen) {
		nsl := make([]T, newlen, newlen*2)
		copy(nsl, sl)
		return nsl
	}
	return sl
}

// Put a value into an index of a slice in place, growing the slice as needed
func SlicePut[T any](sl *[]T, idx int64, val T) {
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]=val
}

// Increment the value stored in the specified index, growing the slice as needed
func SliceInc(sl *[]int64, idx int64) {
	if sl == nil {
		return
	}
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]++
}

// Increment the value in the specified index, but do nothing if idx >= max or slice is nil
func SliceIncMax(sl *[]int64, idx int64, max int64) {
	if sl == nil  || idx >= max {
		return
	}
	*sl = GrowSlice(*sl, idx+1)
	(*sl)[idx]++
}

// Build histogram of all types of pair distances below maxdist. r must point to a .pairs file, and w will output a tab-separated file of counts with this format:
// distance
// paired self trans selfTrans pairedTrans
// pairedIn selfIn transIn selfTransIn pairedTransIn
// pairedOut selfOut transOut selfTransOut pairedTransOut
// pairedMatched selfMatched transMatched selfTransMatched pairedTransMatched
func Run(maxdist int64, r io.Reader, w io.Writer) error {
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
			SliceIncMax(&transcounts, dist, maxdist)
			SliceIncMax(transfacecountp, dist, maxdist)
			if p.Read1.Parent == p.Read2.Parent {
				SliceIncMax(&selftranscounts, dist, maxdist)
				SliceIncMax(selftransfacecountp, dist, maxdist)
			} else {
				SliceIncMax(&pairtranscounts, dist, maxdist)
				SliceIncMax(pairtransfacecountp, dist, maxdist)
			}
			continue
		}
		if p.Read1.Parent == p.Read2.Parent {
			SliceIncMax(&selfcounts, dist, maxdist)
			SliceIncMax(selffacecountp, dist, maxdist)
			continue
		}
		SliceIncMax(&paircounts, dist, maxdist)
		SliceIncMax(pairfacecountp, dist, maxdist)
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
