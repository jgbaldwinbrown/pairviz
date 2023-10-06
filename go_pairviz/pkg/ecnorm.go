package pairviz

import (
	"bufio"
	"flag"
	"regexp"
	"encoding/json"
	"os"
	"fmt"
	"io"
	"compress/gzip"
)

func Close(args ...any) error {
	var err error
	for _, a := range args {
		if c, ok := a.(io.Closer); ok {
			e := c.Close()
			if err == nil {
				err = e
			}
		}
	}
	return err
}

type GzReader struct {
	r *os.File
	gr *gzip.Reader
}

func (g *GzReader) Read(p []byte) (n int, err error) {
	return g.gr.Read(p)
}

func (g *GzReader) Close() error {
	return Close(g.gr, g.r)
}

func OpenGz(path string) (*GzReader, error) {
	r, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	gr, e := gzip.NewReader(r)
	if e != nil {
		defer Must(r.Close())
		return nil, e
	}
	return &GzReader{r, gr}, nil
}

var gzre = regexp.MustCompile(`\.gz$`)

func OpenMaybeGz(path string) (io.ReadCloser, error) {
	if gzre.MatchString(path) {
		return OpenGz(path)
	}
	return os.Open(path)
}

func ParsePairvizOut(r io.Reader) *Iterator[JsonOutStat] {
	return &Iterator[JsonOutStat]{Iteratef: func(yield func(JsonOutStat) error) error {
		dec := json.NewDecoder(r)
		var j JsonOutStat
		for err := dec.Decode(&j); err != io.EOF; err = dec.Decode(&j) {
			if err != nil {
				return err
			}
			if err = yield(j); err != nil {
				return err
			}
		}
		return nil
	}}
}

func AccumStats(sums *JsonOutStat, x JsonOutStat) {
	sums.TargetHits += x.TargetHits
	sums.AltHits += x.AltHits
	sums.TargetProp += x.TargetProp
	sums.AltProp += x.AltProp
	sums.TargetPropGoodBad += x.TargetPropGoodBad
	sums.TargetPropGood += x.TargetPropGood
	sums.TargetPropTotal += x.TargetPropTotal
	sums.TargetFpkm += x.TargetFpkm
	sums.AltFpkm += x.AltFpkm
	sums.TargetFpkmProp += x.TargetFpkmProp
	sums.AltFpkmProp += x.AltFpkmProp
	sums.AltOvlHits += x.AltOvlHits
	sums.AltNonOvlHits += x.AltNonOvlHits
	sums.AltOvlProp += x.AltOvlProp
	sums.AltNonOvlProp += x.AltNonOvlProp
	sums.AltOvlFpkm += x.AltOvlFpkm
	sums.AltNonOvlFpkm += x.AltNonOvlFpkm
	sums.AltOvlFpkmProp += x.AltOvlFpkmProp
	sums.AltNonOvlFpkmProp += x.AltNonOvlFpkmProp
}

func DivCount(sums JsonOutStat, count float64) JsonOutStat {
	out := sums
	out.TargetHits = sums.TargetHits / count
	out.AltHits = sums.AltHits / count
	out.TargetProp = sums.TargetProp / count
	out.AltProp = sums.AltProp / count
	out.TargetPropGoodBad = sums.TargetPropGoodBad / count
	out.TargetPropGood = sums.TargetPropGood / count
	out.TargetPropTotal = sums.TargetPropTotal / count
	out.TargetFpkm = sums.TargetFpkm / count
	out.AltFpkm = sums.AltFpkm / count
	out.TargetFpkmProp = sums.TargetFpkmProp / count
	out.AltFpkmProp = sums.AltFpkmProp / count
	out.AltOvlHits = sums.AltOvlHits / count
	out.AltNonOvlHits = sums.AltNonOvlHits / count
	out.AltOvlProp = sums.AltOvlProp / count
	out.AltNonOvlProp = sums.AltNonOvlProp / count
	out.AltOvlFpkm = sums.AltOvlFpkm / count
	out.AltNonOvlFpkm = sums.AltNonOvlFpkm / count
	out.AltOvlFpkmProp = sums.AltOvlFpkmProp / count
	out.AltNonOvlFpkmProp = sums.AltNonOvlFpkmProp / count
	return out
}

func GetControlStatMeans(controlChr string, it Iter[JsonOutStat]) (JsonOutStat, error) {
	var sums JsonOutStat
	var count int64

	err := it.Iterate(func(j JsonOutStat) error {
		if j.Chr == controlChr {
			AccumStats(&sums, j)
			count++
		}
		return nil
	})
	if err != nil {
		return sums, err
	}

	return DivCount(sums, float64(count)), nil
}

func SubtractControlStat(x JsonOutStat, control JsonOutStat) JsonOutStat {
	out := x
	out.TargetHits = x.TargetHits - control.TargetHits
	out.AltHits = x.AltHits - control.AltHits
	out.TargetProp = x.TargetProp - control.TargetProp
	out.AltProp = x.AltProp - control.AltProp
	out.TargetPropGoodBad = x.TargetPropGoodBad - control.TargetPropGoodBad
	out.TargetPropGood = x.TargetPropGood - control.TargetPropGood
	out.TargetPropTotal = x.TargetPropTotal - control.TargetPropTotal
	out.TargetFpkm = x.TargetFpkm - control.TargetFpkm
	out.AltFpkm = x.AltFpkm - control.AltFpkm
	out.TargetFpkmProp = x.TargetFpkmProp - control.TargetFpkmProp
	out.AltFpkmProp = x.AltFpkmProp - control.AltFpkmProp
	out.AltOvlHits = x.AltOvlHits - control.AltOvlHits
	out.AltNonOvlHits = x.AltNonOvlHits - control.AltNonOvlHits
	out.AltOvlProp = x.AltOvlProp - control.AltOvlProp
	out.AltNonOvlProp = x.AltNonOvlProp - control.AltNonOvlProp
	out.AltOvlFpkm = x.AltOvlFpkm - control.AltOvlFpkm
	out.AltNonOvlFpkm = x.AltNonOvlFpkm - control.AltNonOvlFpkm
	out.AltOvlFpkmProp = x.AltOvlFpkmProp - control.AltOvlFpkmProp
	out.AltNonOvlFpkmProp = x.AltNonOvlFpkmProp - control.AltNonOvlFpkmProp
	return out
}

func SubtractControlStatAll(it Iter[JsonOutStat], control JsonOutStat) *Iterator[JsonOutStat] {
	return &Iterator[JsonOutStat]{Iteratef: func(yield func(JsonOutStat) error) error {
		return it.Iterate(func(x JsonOutStat) error {
			j := SubtractControlStat(x, control)
			return yield(j)
		})
	}}
}

func FullSubtractControl() {
	controlChr := flag.String("c", "", "Chromosome to use as control (required)")
	inpath := flag.String("i", "", "Input path (required")
	flag.Parse()
	if *controlChr == "" {
		panic(fmt.Errorf("missing -c"))
	}
	if *inpath == "" {
		panic(fmt.Errorf("missing -i"))
	}

	r, e := OpenMaybeGz(*inpath)
	Must(e)

	it := ParsePairvizOut(r)
	cmean, e := GetControlStatMeans(*controlChr, it)
	Must(r.Close())
	Must(e)

	r, e = OpenMaybeGz(*inpath)
	Must(e)
	defer Must(r.Close())

	it = ParsePairvizOut(r)
	subit := SubtractControlStatAll(it, cmean)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	enc := json.NewEncoder(w)
	err := subit.Iterate(func(j JsonOutStat) error {
		return enc.Encode(j)
	})
	Must(err)
}
