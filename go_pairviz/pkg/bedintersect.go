package pairviz

import (
	"iter"
	"io"
	"encoding/csv"
	"os"
	"flag"
	"log"

	"github.com/jgbaldwinbrown/span"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/zfile"
	"github.com/jgbaldwinbrown/iterh"
)

type Span struct {
	fastats.Span
}

func (s Span) Left() int64 {
	return s.Start
}

func (s Span) Right() int64 {
	return s.End
}

func ChrSpanToSet[CS fastats.ChrSpanner](it iter.Seq[CS]) map[string]*span.OrderedSet[Span, int64] {
	m1 := map[string][]Span{}
	m := map[string]*span.OrderedSet[Span, int64]{}
	for b := range it {
		m1[b.SpanChr()] = append(m1[b.SpanChr()], Span{fastats.Span{b.SpanStart(), b.SpanEnd()}})
	}
	for chr, vals := range m1 {
		m[chr] = span.NewOrderedSet(vals)
	}

	return m
}

func BedfileToSet(path string) (set map[string]*span.OrderedSet[Span, int64], err error) {
	r, e := zfile.Open(path)
	if e != nil {
		return nil, e
	}
	defer func() {
		e := r.Close()
		if err == nil {
			err = e
		}
	}()
	bed, ep := iterh.BreakWithError(fastats.ParseBedFlat(r))
	set = ChrSpanToSet(bed)
	return set, *ep
}

func TouchSet(set map[string]*span.OrderedSet[Span, int64], r Read) bool {
	cset, ok := set[r.Chrom]
	if !ok || cset == nil {
		return false
	}
	return cset.Touching(Span{fastats.Span{r.Left(), r.Right()}})
}

func FilterPairvizBadBed(bads map[string]*span.OrderedSet[Span, int64], r io.Reader, w io.Writer) (err error) {
	cw := csv.NewWriter(w)
	cw.Comma = '\t'
	defer func() {
		cw.Flush()
		if err == nil {
			err = cw.Error()
		}
	}()

	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	cr.Comma = '\t'
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return e
		}
		p, ok := ParsePair(l)
		if !ok || !IsAPair(l) || !CheckGood(l) {
			if e := cw.Write(l); e != nil {
				return e
			}
			continue
		}
		if !TouchSet(bads, p.Read1) && !TouchSet(bads, p.Read2) {
			if e := cw.Write(l); e != nil {
				return e
			}
		}
	}
	return nil
}

type filterPairsBadBedFlags struct {
	Bed string
}

func FullFilterPairsBadBed() {
	var f filterPairsBadBedFlags
	flag.StringVar(&f.Bed, "b", "", "Bed file to filter out (required)")
	flag.Parse()
	if f.Bed == "" {
		log.Fatal("missing -b")
	}

	b, e := BedfileToSet(f.Bed)
	if e != nil {
		log.Fatal(e)
	}
	if e := FilterPairvizBadBed(b, os.Stdin, os.Stdout); e != nil {
		log.Fatal(e)
	}
}
