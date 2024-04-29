package windif

import (
	"fmt"
	"encoding/json"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/csvh"
	"bufio"
	"io"
	"os"
	"log"
	"strings"
)

func ReadLines(r io.Reader) ([]string, error) {
	b, e := io.ReadAll(r)
	if e != nil {
		return nil, e
	}
	return strings.Split(strings.TrimRight(string(b), "\n"), "\n"), nil
}

type Winpair struct {
	Fa1 fastats.FaEntry
	Fa2 fastats.FaEntry
}

func WriteFaPath(path string, fa []fastats.FaEntry) (err error) {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer func() {
		e := w.Close()
		if err == nil {
			err = e
		}
	}()

	return fastats.WriteFaEntries(w, fa...)
}

func Runs(wp Winpair) []fastats.Span {
	var out []fastats.Span
	var start int64 = 0
	var end int64 = 0
	for i, c := range []byte(wp.Fa1.Seq) {
		end = int64(i)
		d := wp.Fa2.Seq[i]
		if c != d {
			if start != end {
				out = append(out, fastats.Span{Start: start, End: end})
			}
			start = end + 1
		}
	}
	if start < int64(len(wp.Fa1.Seq)) {
		out = append(out, fastats.Span{Start: start, End: end})
	}
	return out
}

func CountRuns(runs []fastats.Span, min int64) int64 {
	var count int64
	for _, r := range runs {
		l := r.End - r.Start
		if l >= min {
			count += l - min + 1
		}
	}
	return count
}

func RunsPerBp(wp Winpair, min int64) float64 {
	runs := Runs(wp)
	count := CountRuns(runs, min)
	return float64(count) / float64(len(wp.Fa1.Seq))
}

type WinpairStat struct {
	TripletsPerBp float64
	RunsPerBp float64
	MummerMatchBp int64
	BlastBitscore int64
}

func WriteStats(w io.Writer, wss ...WinpairStat) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	for _, ws := range wss {
		e := enc.Encode(ws)
		if e != nil {
			return e
		}
	}
	return nil
}

func WinpairStats(wp Winpair) (WinpairStat, error) {
	var s WinpairStat
	h := func(e error) (WinpairStat, error) {
		return WinpairStat{}, fmt.Errorf("WinpairStats: %w", e)
	}
	var e error
	log.Println("starting runs")
	s.RunsPerBp = RunsPerBp(wp, 100)
	log.Println("finished runs")
	s.TripletsPerBp = TripletsPerBp(wp, 100, 1)
	log.Println("finished triplets")
	s.MummerMatchBp, e = MummerMatchBp(wp)
	if e != nil {
		return h(e)
	}
	log.Println("finished mummer")
	s.BlastBitscore, e = BestBitscore(wp)
	log.Println("finished blast")
	return s, e
}

func Run() {
	paths, e := ReadLines(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	it := Winpairs(paths)

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()
	e = it.Iterate(func(wp Winpair) error {
		stat, e := WinpairStats(wp)
		if e != nil {
			return e
		}
		e = WriteStats(w, stats...)
		if e != nil {
			return e
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}

}

// temp/mummer/melref2sim/mummer_out_melref2sim.delta: raw_data/individual_strain_assemblies/iso1/dmel_rel6_sorted_ne>
//         mkdir -p `dirname $@`
//         nucmer -l 100 -prefix temp/mummer/melref2sim/mummer_out_melref2sim $^
// temp/mummer/melref2sim/mummer_out_melref2sim.rq.delta: temp/mummer/melref2sim/mummer_out_melref2sim.delta
//         mkdir -p `dirname $@`
//         delta-filter -i 95 -r -q $< > $@
// temp/mummer/melref2sim/mummer_out_melref2sim.coords: temp/mummer/melref2sim/mummer_out_melref2sim.rq.delta
//         show-coords -o -l -r $< > $@
// temp/mummer/melref2sim/mummer_out_melref2sim.coords.tsv: temp/mummer/melref2sim/mummer_out_melref2sim.rq.delta
//         show-coords -o -l -r -T $< > $@
