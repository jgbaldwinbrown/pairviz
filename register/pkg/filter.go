package register

import (
	"log"
	"fmt"
	"os"
	"io"
	"flag"
	"encoding/json"
	"encoding/csv"
)

type ReadType int

const (
	TypeUnknown ReadType = iota
	PairType
	SelfType
	TransType
)

type FilterSet struct {
	MaxDist int64
	MinDist int64
	Faces []Facing
	ReadTypes []ReadType
}

func (f FilterSet) Check(dist int64, face Facing, rtype ReadType) bool {
	if dist < f.MinDist {
		return false
	}
	if dist > f.MaxDist {
		return false
	}

	ok := false
	for _, fface := range f.Faces {
		if face == fface {
			ok = true
			break
		}
	}
	if !ok {
		return false
	}

	ok = false
	for _, ftype := range f.ReadTypes {
		if rtype == ftype {
			ok = true
			break
		}
	}
	if !ok {
		return false
	}

	return true
}

type FilterArgs struct {
	FilterSets []FilterSet
}

func GetFilterArgsFromReader(r io.Reader) (FilterArgs, error) {
	h := handle("GetFilterArgsFromReader: %w")
	var args FilterArgs

	dec := json.NewDecoder(r)
	e := dec.Decode(&args)
	if e != nil { return args, h(e) }

	return args, nil
}

func GetFilterArgsFromPath(path string) (FilterArgs, error) {
	h := handle("GetFilterArgsFromPath: %w")

	r, e := os.Open(path)
	if e != nil { return FilterArgs{}, h(e) }
	defer r.Close()

	return GetFilterArgsFromReader(r)
}

func GetFilterArgs() (FilterArgs, error) {
	argpathp := flag.String("a", "", "argument file")
	flag.Parse()
	if *argpathp == "" {
		return FilterArgs{}, fmt.Errorf("missing -a")
	}

	args, e := GetFilterArgsFromPath(*argpathp)
	if e != nil { return args, e }

	return args, nil
}

func RunFilter(r io.Reader, w io.Writer, args FilterArgs) error {
	h := handle("Run: %w")
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.Comma = rune('\t')
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		p, ok := ParsePair(line)
		if !ok { log.Printf("pair %v not ok\n", p); continue }

		if !p.Read1.Ok || !p.Read2.Ok { continue }

		dist := Abs(p.Read2.Pos - p.Read1.Pos)
		face := p.Face()
		rtype := p.ReadType()

		for _, filt := range args.FilterSets {
			if filt.Check(dist, face, rtype) {
				cw.Write(line)
				break
			}
		}
	}

	return nil
}
