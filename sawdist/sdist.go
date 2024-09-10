package main

import (
	"io"
	"fmt"
	"sort"
	"log"
	"errors"
	"flag"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"iter"
	"github.com/jgbaldwinbrown/csvh"
	"strconv"
)

// All the fields of the input .bed file, plus the parsed PairPropFpkm
type PairFields struct {
	PairPropFpkm float64
	Fields []string
}

// Remove the first line from a file
func TrimHead(r io.Reader) error {
	buf := make([]byte, 1)
	for n, e := r.Read(buf); e != io.EOF; n, e = r.Read(buf) {
		if n < 1 {
			continue
		}
		if e != nil {
			return e
		}
		if buf[0] == '\n' {
			break
		}
	}
	return nil
}

// Collect all the values from an iterator and break if there's an error
func CollectErr[T any](it iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for val, e := range it {
		if e != nil {
			return nil, e
		}
		out = append(out, val)
	}
	return out, nil
}

// Parse a bed file, assuming fields[13] is the PairPropFpkm (a float); also sort by position
func GetBed(path string) ([]fastats.BedEntry[PairFields], error) {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	if e := TrimHead(r); e != nil {
		return nil, e
	}

	bed, e := CollectErr[fastats.BedEntry[PairFields]](fastats.ParseBed(r, func(fields []string) (PairFields, error) {
		var pf PairFields
		pf.PairPropFpkm, e = strconv.ParseFloat(fields[13], 64)
		if e != nil {
			return pf, e
		}
		pf.Fields = make([]string, len(fields))
		copy(pf.Fields, fields)
		return pf, nil
	}))
	if e != nil {
		return nil, e
	}

	sort.Slice(bed, func (i, j int) bool {
		bi := bed[i]
		bj := bed[j]
		if bi.Chr < bj.Chr {
			return true
		} else if bi.Chr > bj.Chr {
			return false
		} else if bi.Start < bj.Start {
			return true
		} else if bi.Start > bj.Start {
			return false
		} else if bi.End < bj.End {
			return true
		}
		return false
	})

	return bed, nil
}

// Combined information for one bed location from pure species, hybrid, and sawamura crosses
type JoinedBedEntry struct {
	Pure fastats.BedEntry[PairFields]
	Hyb fastats.BedEntry[PairFields]
	Saw fastats.BedEntry[PairFields]
}

// Convert a slice of bed entries to a map of the same entries, with positions as the keys (for combining)
func ToMap[T any](b []fastats.BedEntry[T]) map[fastats.ChrSpan]*fastats.BedEntry[T] {
	m := make(map[fastats.ChrSpan]*fastats.BedEntry[T], len(b))
	for i, be := range b {
		m[be.ChrSpan] = &b[i]
	}
	return m
}

// Join three beds by position
func Join(pure, hyb, saw []fastats.BedEntry[PairFields]) []JoinedBedEntry {
	hm := ToMap[PairFields](hyb)
	sm := ToMap[PairFields](saw)
	var out []JoinedBedEntry
	for _, p := range pure {
		h, hok := hm[p.ChrSpan]
		s, sok := sm[p.ChrSpan]
		if hok && sok {
			out = append(out, JoinedBedEntry{p, *h, *s})
		}
	}
	return out
}

// Find the new value for Sawamura, scaled by the control values, where pure species = 1 and hybrid = 0
func Dist(b JoinedBedEntry) fastats.BedEntry[float64] {
	full := b.Pure.Fields.PairPropFpkm - b.Hyb.Fields.PairPropFpkm
	dist := b.Saw.Fields.PairPropFpkm - b.Hyb.Fields.PairPropFpkm
	return fastats.BedEntry[float64]{ChrSpan: b.Pure.ChrSpan, Fields: dist / full}
}

// Run Dist for an entire bed file
func Dists(bs ...JoinedBedEntry) []fastats.BedEntry[float64] {
	out := make([]fastats.BedEntry[float64], 0, len(bs))
	for _, b := range bs {
		out = append(out, Dist(b))
	}
	return out
}

func Run(purepath, hybpath, sawpath, outpath string) (err error) {
	p, e := GetBed(purepath)
	if e != nil {
		return e
	}
	h, e := GetBed(hybpath)
	if e != nil {
		return e
	}
	s, e := GetBed(sawpath)
	if e != nil {
		return e
	}

	j := Join(p, h, s)
	dists := Dists(j...)

	out, e := csvh.CreateMaybeGz(outpath)
	if e != nil {
		return e
	}
	defer func() { csvh.DeferE(&err, out.Close()) }()
	for _, d := range dists {
		if _, e := fmt.Fprintf(out, "%v\t%v\t%v\t%v\n", d.Chr, d.Start, d.End, d.Fields); e != nil {
			return e
		}
	}
	return nil
}

type Flags struct {
	Pure string
	Hyb string
	Saw string
	Out string
}

func main() {
	var f Flags
	flag.StringVar(&f.Pure, "p", "", "path to bed file with pure species pairing values")
	flag.StringVar(&f.Hyb, "h", "", "path to bed file with hybrid pairing values")
	flag.StringVar(&f.Saw, "s", "", "path to bed file with sawamura pairing values")
	flag.StringVar(&f.Out, "o", "", "path to write output")
	flag.Parse()
	if f.Pure == "" || f.Hyb == "" || f.Saw == "" || f.Out == "" {
		log.Fatal(errors.New("Missing argument"))
	}
	e := Run(f.Pure, f.Hyb, f.Saw, f.Out)
	if e != nil {
		log.Fatal(e)
	}
}
