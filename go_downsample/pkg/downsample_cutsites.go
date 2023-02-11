package dsamp

import (
	"fmt"
	"math/rand"
	"os"
	"compress/gzip"
	"io"
	"bufio"
)

func CountBedLines(r io.Reader) (int, error) {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	i := 0
	for s.Scan() {
		i++
	}
	return i, nil
}

func CountBedLinesGzfile(path string) (int, error) {
	r, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	gr, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}
	defer gr.Close()

	return CountBedLines(gr)
}

func SubsetBed(r io.Reader, w io.Writer, seed int64, prop float64) error {
	rd := rand.New(rand.NewSource(seed))
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for s.Scan() {
		roll := rd.Float64()
		if roll <= prop {
			fmt.Fprintln(bw, s.Text())
		}
	}
	return nil
}

func CountBedLinesGzfiles(paths ...string) ([]int, error) {
	var counts []int
	for _, path := range paths {
		count, err := CountBedLinesGzfile(path)
		if err != nil {
			return counts, err
		}
		counts = append(counts, count)
	}
	return counts, nil
}

func SubsetBedGzpath(arg SubsetArgs) error {
	r, err := os.Open(arg.Inpath)
	if err != nil {
		return err
	}
	defer r.Close()
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	w, err := os.Create(arg.Outpath)
	if err != nil {
		return err
	}
	defer w.Close()
	gzw := gzip.NewWriter(w)
	defer gzw.Close()
	bw := bufio.NewWriter(gzw)
	defer bw.Flush()

	return SubsetBed(gzr, gzw, arg.Seed, arg.Prop)
}

func SubsetBedGzpaths(args ...SubsetArgs) error {
	for _, arg := range args {
		err := SubsetBedGzpath(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunDownsampleCutSites() {
	args := GetArgs()
	var inpaths []string
	for _, set := range args.IoSets {
		inpaths = append(inpaths, set.Inpath)
	}
	counts, err := CountBedLinesGzfiles(inpaths...)
	if err != nil {
		panic(err)
	}

	countw, err := os.Create(args.Countspath)
	if err != nil {
		panic(err)
	}
	defer countw.Close()
	PrintPathCounts(countw, inpaths, counts)

	props := GetLowestProps(counts)

	subargs := MakeSubsetArgs(args, props)
	err = SubsetBedGzpaths(subargs...)
	if err != nil {
		panic(err)
	}
}
