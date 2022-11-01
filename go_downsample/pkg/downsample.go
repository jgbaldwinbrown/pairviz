package dsamp

import (
	"encoding/json"
	"fmt"
	"os"
	"compress/gzip"
	"bufio"
	"io"
	"math/rand"
)

func SubsetPairvizUnique(r io.Reader, w io.Writer, seed int64, prop float64) error {
	rd := rand.New(rand.NewSource(seed))
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	var line []string

	for s.Scan() {
		if IsComment(s.Text()) {
			fmt.Fprintln(bw, s.Text())
		}
		if IsUnique(s.Text(), line) {
			roll := rd.Float64()
			if roll <= prop {
				fmt.Fprintln(bw, s.Text())
			}
		}
	}
	return nil
}

type SubsetArgs struct {
	Inpath string
	Outpath string
	Seed int64
	Prop float64
}

func SubsetPairvizUniqueGzpath(arg SubsetArgs) error {
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

	return SubsetPairvizUnique(gzr, gzw, arg.Seed, arg.Prop)
}

func SubsetPairvizUniqueGzpaths(args ...SubsetArgs) error {
	for _, arg := range args {
		err := SubsetPairvizUniqueGzpath(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

type IoSet struct {
	Inpath string
	Outpath string
	Seed int64
}

type DownsampleArgs struct {
	IoSets []IoSet
	Countspath string
}

func GetArgs() DownsampleArgs {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var args DownsampleArgs
	err = json.Unmarshal(bytes, &args)
	if err != nil {
		panic(err)
	}

	return args
}

func GetLowestProps(counts []int) []float64 {
	lowest := 1000000000000000
	for _, count := range counts {
		if lowest > count {
			lowest = count
		}
	}

	var props []float64
	for _, count := range counts {
		props = append(props, float64(lowest) / float64(count))
	}
	return props
}

func RunDownsample() {
	args := GetArgs()
	var inpaths []string
	for _, set := range args.IoSets {
		inpaths = append(inpaths, set.Inpath)
	}
	counts, err := CountPairtoolsUniqueGzfiles(inpaths...)
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

	var subargs []SubsetArgs
	for i, set := range args.IoSets {
		subargs = append(subargs,
			SubsetArgs{
				Inpath: set.Inpath,
				Outpath: set.Outpath,
				Seed: set.Seed,
				Prop: props[i],
		})
	}

	err = SubsetPairvizUniqueGzpaths(subargs...)
	if err != nil {
		panic(err)
	}
}
