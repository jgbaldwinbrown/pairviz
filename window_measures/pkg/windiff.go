package windif

import (
	"flag"
	"fmt"
	"encoding/json"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/csvh"
	// "bufio"
	"io"
	"os"
	"log"
	"strings"
	"github.com/jgbaldwinbrown/parallel_ordered"
	"golang.org/x/sync/errgroup"
	"iter"
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
		if !BaseMatch(c, d) {
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

func PerfectMatches(wp Winpair) int64 {
	var sum int64 = 0
	for i, c := range []byte(wp.Fa1.Seq) {
		d := wp.Fa2.Seq[i]
		if BaseMatch(c, d) {
			sum++
		}
	}
	return sum
}

func Identity(wp Winpair) float64 {
	ms := PerfectMatches(wp)
	return float64(ms) / float64(len(wp.Fa1.Seq))
}

type WinpairStat struct {
	TripletsPerBp Float
	RunsPerBp Float
	MummerMatchBp int64
	BlastBitscore int64
	Identity Float
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
	s.RunsPerBp.F = RunsPerBp(wp, 100)
	s.TripletsPerBp.F = TripletsPerBp(wp, 100, 1)
	s.Identity.F = Identity(wp)

	s.MummerMatchBp, e = MummerMatchBp(wp)
	if e != nil {
		log.Println(h(e))
		s.MummerMatchBp = -1
	}
	s.BlastBitscore, e = BestBitscore(wp)
	if e != nil {
		log.Println(h(e))
		s.BlastBitscore = -1
	}
	return s, nil
}

type WindiffFlags struct {
	Threads int
}

func Run() {
	var f WindiffFlags
	flag.IntVar(&f.Threads, "t", -1, "Threads to use (default infinite)")
	flag.Parse()

	paths, e := ReadLines(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	it := Winpairs(paths)

	w := os.Stdout
	// w := bufio.NewWriter(os.Stdout)
	// defer func() {
	// 	e := w.Flush()
	// 	if e != nil {
	// 		log.Fatal(e)
	// 	}
	// }()

	o := po.NewOrderer[po.IndexedVal[WinpairStat]](1024)
	var g errgroup.Group
	if f.Threads > 0 {
		g.SetLimit(f.Threads)
	}

	writeI := 0
	writeErr := make(chan error)
	go func() {
		for val, ok := o.Read(); ok; val, ok = o.Read() {
			e := WriteStats(w, val.Val)
			if e != nil {
				writeErr <- e
				return
			}
			writeI++
		}
		writeErr <- nil
		return
	}()

	i := 0
	for wp, e := range it {
		if e != nil {
			log.Fatal(e)
		}
		j := i
		g.Go(func() error {
			k := j
			stat, e := WinpairStats(wp)
			if e != nil {
				return e
			}
			o.Write(po.IndexedVal[WinpairStat]{Val: stat, I: k})
			return nil
		})
		i++
	}

	if e := g.Wait(); e != nil {
		log.Fatal(e)
	}
	o.Close()

	if e := <-writeErr; e != nil {
		log.Fatal(e)
	}
}
