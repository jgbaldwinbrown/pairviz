package pairviz

import (
	"iter"
	"math"
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
		defer func() { Must(r.Close()) }()
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

func ParsePairvizOut(r io.Reader) iter.Seq2[JsonOutStat, error] {
	return func(yield func(JsonOutStat, error) bool) {
		dec := json.NewDecoder(r)
		var j JsonOutStat
		for err := dec.Decode(&j); err != io.EOF; err = dec.Decode(&j) {
			if ok := yield(j, err); !ok {
				return
			}
		}
	}
}

func IsInfOrNaN(j JsonFloat) bool {
	f := float64(j)
	if math.IsInf(f, 0) {
		return true
	}
	return math.IsNaN(f)
}

func AccumStat(sum *JsonFloat, count *JsonFloat, x JsonFloat) {
	if IsInfOrNaN(x) {
		return
	}
	*sum += x
	(*count)++
}

func AccumStats(sums *JsonOutStat, counts *JsonOutStat, x JsonOutStat) {
	AccumStat(&sums.TargetHits, &counts.TargetHits, x.TargetHits)
	AccumStat(&sums.AltHits, &counts.AltHits, x.AltHits)
	AccumStat(&sums.TargetProp, &counts.TargetProp, x.TargetProp)
	AccumStat(&sums.AltProp, &counts.AltProp, x.AltProp)
	AccumStat(&sums.TargetPropGoodBad, &counts.TargetPropGoodBad, x.TargetPropGoodBad)
	AccumStat(&sums.TargetPropGood, &counts.TargetPropGood, x.TargetPropGood)
	AccumStat(&sums.TargetPropTotal, &counts.TargetPropTotal, x.TargetPropTotal)
	AccumStat(&sums.TargetFpkm, &counts.TargetFpkm, x.TargetFpkm)
	AccumStat(&sums.AltFpkm, &counts.AltFpkm, x.AltFpkm)
	AccumStat(&sums.TargetFpkmProp, &counts.TargetFpkmProp, x.TargetFpkmProp)
	AccumStat(&sums.AltFpkmProp, &counts.AltFpkmProp, x.AltFpkmProp)
	AccumStat(&sums.AltOvlHits, &counts.AltOvlHits, x.AltOvlHits)
	AccumStat(&sums.AltNonOvlHits, &counts.AltNonOvlHits, x.AltNonOvlHits)
	AccumStat(&sums.AltOvlProp, &counts.AltOvlProp, x.AltOvlProp)
	AccumStat(&sums.AltNonOvlProp, &counts.AltNonOvlProp, x.AltNonOvlProp)
	AccumStat(&sums.AltOvlFpkm, &counts.AltOvlFpkm, x.AltOvlFpkm)
	AccumStat(&sums.AltNonOvlFpkm, &counts.AltNonOvlFpkm, x.AltNonOvlFpkm)
	AccumStat(&sums.AltOvlFpkmProp, &counts.AltOvlFpkmProp, x.AltOvlFpkmProp)
	AccumStat(&sums.AltNonOvlFpkmProp, &counts.AltNonOvlFpkmProp, x.AltNonOvlFpkmProp)
}

func DivCount(sums JsonOutStat, count float64) JsonOutStat {
	jcount := JsonFloat(count)

	out := sums
	out.TargetHits = sums.TargetHits / jcount
	out.AltHits = sums.AltHits / jcount
	out.TargetProp = sums.TargetProp / jcount
	out.AltProp = sums.AltProp / jcount
	out.TargetPropGoodBad = sums.TargetPropGoodBad / jcount
	out.TargetPropGood = sums.TargetPropGood / jcount
	out.TargetPropTotal = sums.TargetPropTotal / jcount
	out.TargetFpkm = sums.TargetFpkm / jcount
	out.AltFpkm = sums.AltFpkm / jcount
	out.TargetFpkmProp = sums.TargetFpkmProp / jcount
	out.AltFpkmProp = sums.AltFpkmProp / jcount
	out.AltOvlHits = sums.AltOvlHits / jcount
	out.AltNonOvlHits = sums.AltNonOvlHits / jcount
	out.AltOvlProp = sums.AltOvlProp / jcount
	out.AltNonOvlProp = sums.AltNonOvlProp / jcount
	out.AltOvlFpkm = sums.AltOvlFpkm / jcount
	out.AltNonOvlFpkm = sums.AltNonOvlFpkm / jcount
	out.AltOvlFpkmProp = sums.AltOvlFpkmProp / jcount
	out.AltNonOvlFpkmProp = sums.AltNonOvlFpkmProp / jcount
	return out
}

func DivCounts(sums JsonOutStat, counts JsonOutStat) JsonOutStat {
	out := sums
	out.TargetHits = sums.TargetHits / counts.TargetHits
	out.AltHits = sums.AltHits / counts.AltHits
	out.TargetProp = sums.TargetProp / counts.TargetProp
	out.AltProp = sums.AltProp / counts.AltProp
	out.TargetPropGoodBad = sums.TargetPropGoodBad / counts.TargetPropGoodBad
	out.TargetPropGood = sums.TargetPropGood / counts.TargetPropGood
	out.TargetPropTotal = sums.TargetPropTotal / counts.TargetPropTotal
	out.TargetFpkm = sums.TargetFpkm / counts.TargetFpkm
	out.AltFpkm = sums.AltFpkm / counts.AltFpkm
	out.TargetFpkmProp = sums.TargetFpkmProp / counts.TargetFpkmProp
	out.AltFpkmProp = sums.AltFpkmProp / counts.AltFpkmProp
	out.AltOvlHits = sums.AltOvlHits / counts.AltOvlHits
	out.AltNonOvlHits = sums.AltNonOvlHits / counts.AltNonOvlHits
	out.AltOvlProp = sums.AltOvlProp / counts.AltOvlProp
	out.AltNonOvlProp = sums.AltNonOvlProp / counts.AltNonOvlProp
	out.AltOvlFpkm = sums.AltOvlFpkm / counts.AltOvlFpkm
	out.AltNonOvlFpkm = sums.AltNonOvlFpkm / counts.AltNonOvlFpkm
	out.AltOvlFpkmProp = sums.AltOvlFpkmProp / counts.AltOvlFpkmProp
	out.AltNonOvlFpkmProp = sums.AltNonOvlFpkmProp / counts.AltNonOvlFpkmProp
	return out
}

func GetControlStatMeans(controlChr string, it iter.Seq2[JsonOutStat, error]) (control, exp JsonOutStat, err error) {
	var sums JsonOutStat
	var counts JsonOutStat

	var expsums JsonOutStat
	var expcounts JsonOutStat

	for j, err := range it {
		if err != nil {
			return JsonOutStat{}, JsonOutStat{}, err
		}
		if j.Chr == controlChr {
			AccumStats(&sums, &counts, j)
		} else {
			AccumStats(&expsums, &expcounts, j)
		}
	}

	return DivCounts(sums, counts), DivCounts(expsums, expcounts), nil
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

func SubtractControlStatAll(it iter.Seq2[JsonOutStat, error], control JsonOutStat) iter.Seq2[JsonOutStat, error] {
	return func(yield func(JsonOutStat, error) bool) {
		for x, err := range it {
			j := SubtractControlStat(x, control)
			if ok := yield(j, err); !ok {
				return
			}
		}
	}
}

func DivideControlAltFpkm(x JsonOutStat, control JsonOutStat) JsonOutStat {
	out := x
	out.TargetHits = x.TargetHits / control.AltFpkm
	out.AltHits = x.AltHits / control.AltFpkm
	out.TargetProp = x.TargetProp / control.AltFpkm
	out.AltProp = x.AltProp / control.AltFpkm
	out.TargetPropGoodBad = x.TargetPropGoodBad / control.AltFpkm
	out.TargetPropGood = x.TargetPropGood / control.AltFpkm
	out.TargetPropTotal = x.TargetPropTotal / control.AltFpkm
	out.TargetFpkm = x.TargetFpkm / control.AltFpkm
	out.AltFpkm = x.AltFpkm / control.AltFpkm
	out.TargetFpkmProp = x.TargetFpkmProp / control.AltFpkm
	out.AltFpkmProp = x.AltFpkmProp / control.AltFpkm
	out.AltOvlHits = x.AltOvlHits / control.AltFpkm
	out.AltNonOvlHits = x.AltNonOvlHits / control.AltFpkm
	out.AltOvlProp = x.AltOvlProp / control.AltFpkm
	out.AltNonOvlProp = x.AltNonOvlProp / control.AltFpkm
	out.AltOvlFpkm = x.AltOvlFpkm / control.AltFpkm
	out.AltNonOvlFpkm = x.AltNonOvlFpkm / control.AltFpkm
	out.AltOvlFpkmProp = x.AltOvlFpkmProp / control.AltFpkm
	out.AltNonOvlFpkmProp = x.AltNonOvlFpkmProp / control.AltFpkm
	return out
}

func DivideControlAltFpkmAll(it iter.Seq2[JsonOutStat, error], control JsonOutStat) iter.Seq2[JsonOutStat, error] {
	return func(yield func(JsonOutStat, error) bool) {
		for x, err := range it {
			j := DivideControlAltFpkm(x, control)
			if ok := yield(j, err); !ok {
				return
			}
		}
	}
}

func WriteMeansPath(path string, control, exp JsonOutStat) error {
	h := func(e error) error {
		return fmt.Errorf("WriteMeansPath: %w", e)
	}

	w, e := os.Create(path)
	if e != nil {
		return h(e)
	}
	defer w.Close()

	enc := json.NewEncoder(w)

	if e = enc.Encode(control); e != nil {
		return h(e)
	}
	if e = enc.Encode(exp); e != nil {
		return h(e)
	}
	return nil
}

func FullSubtractControl() {
	controlChr := flag.String("c", "", "Chromosome to use as control (required)")
	inpath := flag.String("i", "", "Input path (default stdin)")
	outmeanp := flag.String("mo", "", "Path to output means to (default discard)")
	divp := flag.Bool("div", false, "divide all applicable results by the control alt fpkm")
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
	cmean, emean, e := GetControlStatMeans(*controlChr, it)
	Must(r.Close())
	Must(e)

	if *outmeanp != "" {
		WriteMeansPath(*outmeanp, cmean, emean)
	}

	r, e = OpenMaybeGz(*inpath)
	Must(e)
	defer func() { Must(r.Close()) }()

	it = ParsePairvizOut(r)
	var transit iter.Seq2[JsonOutStat, error]
	if *divp {
		transit = DivideControlAltFpkmAll(it, cmean)
	} else {
		transit = SubtractControlStatAll(it, cmean)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	enc := json.NewEncoder(w)
	for j, err := range transit {
		Must(err)
		Must(enc.Encode(j))
	}
}
