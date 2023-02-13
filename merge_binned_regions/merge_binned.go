package main

import (
	"bufio"
	"io"
	"os"
	"fmt"
	"os/exec"
	"flag"
	"log"
	"encoding/csv"
)

func handle(format string) func(...any) error {
	return func(args ...any) error {
		return fmt.Errorf(format, args...)
	}
}

func OpenCsv(path string) (*os.File, *csv.Reader, error) {
	h := handle("OpenCsv: %w")

	r, e := os.Open(path)
	if e != nil { return nil, nil, h(e) }

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	return r, cr, nil
}

func CreateCsv(path string) (*os.File, *csv.Writer, error) {
	h := handle("CreateCsv: %w")

	w, e := os.Create(path)
	if e != nil { return nil, nil, h(e) }

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')

	return w, cw, nil
}

func FindBins(col int, path string) ([]string, error) {
	h := handle("FindBins: %w")

	var bins []string
	set := map[string]struct{}{}

	r, cr, e := OpenCsv(path)
	if e != nil { return nil, h(e) }
	defer r.Close()

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, h(e) }

		if len(line) <= col {
			continue
		}

		if _, ok := set[line[col]]; !ok {
			set[line[col]] = struct{}{}
			bins = append(bins, line[col])
		}
	}

	return bins, nil
}

func SplitBin(col int, bins []string, path, opath string) error {
	h := handle("SplitBin: %w")

	r, cr, e := OpenCsv(path)
	if e != nil { return h(e) }
	defer r.Close()

	w, cw, e := CreateCsv(opath)
	if e != nil { return h(e) }
	defer w.Close()
	defer cw.Flush()

	binset := map[string]struct{}{}
	for _, bin := range bins {
		binset[bin] = struct{}{}
	}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) <= col {
			continue
		}

		if _, ok := binset[line[col]]; ok {
			e = cw.Write(line)
			if e != nil { return h(e) }
		}
	}

	return nil
}

func SplitByBins(col int, bins []string, path, opre string) error {
	h := handle("SplitByBins: %w")

	for _, bin := range bins {
		opath := opre + "_" + bin + ".bed"
		e := SplitBin(col, []string{bin}, path, opath)
		if e != nil { return h(e) }
	}
	return nil
}

func AllBut[T comparable](tolose string, bins []string) []string {
	var out []string
	for _, bin := range bins {
		if tolose != bin {
			out = append(out, bin)
		}
	}
	return out
}

func SplitByBinsBg(col int, bins []string, path, opre string) error {
	h := handle("SplitByBinsBg: %w")

	for _, bin := range bins {
		opath := opre + "_" + bin + "_bg.bed"
		e := SplitBin(col, AllBut[string](bin, bins), path, opath)
		if e != nil { return h(e) }
	}
	return nil
}

func JoinSplit(inpath, outpath string) error {
	h := handle("JoinSplit: %w")

	w, e := os.Create(outpath)
	if e != nil { return h(e) }
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	cmd := exec.Command("bedtools", "merge", "-i", inpath)
	cmd.Stdout = bw
	cmd.Stderr = os.Stderr

	e = cmd.Run()
	if e != nil { return h(e) }

	return nil
}

func JoinSplits(bins []string, opre string) error {
	for _, bin := range bins {
		inpath := opre + "_" + bin + ".bed"
		outpath := opre + "_" + bin + "_joined.bed"
		e := JoinSplit(inpath, outpath)
		if e != nil { return fmt.Errorf("JoinSplits: %w") }
	}
	return nil
}

func JoinSplitsBg(bins []string, opre string) error {
	for _, bin := range bins {
		inpath := opre + "_" + bin + "_bg.bed"
		outpath := opre + "_" + bin + "_bg_joined.bed"
		e := JoinSplit(inpath, outpath)
		if e != nil { return fmt.Errorf("JoinSplits: %w") }
	}
	return nil
}

func GetFasta(fapath, inpath, outpath string) error {
	h := handle("GetFasta: %w")
	cmd := exec.Command("bedtools", "getfasta", "-bed", inpath, "-fo", outpath, "-fi", fapath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e := cmd.Run()
	if e != nil { return h(e) }
	return nil
}

func GetFastas(fapath string, bins []string, opre string) error {
	h := handle("GetFastas: %w")

	for _, bin := range bins {
		inpath := opre + "_" + bin + "_joined.bed"
		outpath := opre + "_" + bin + "_joined.fa"
		e := GetFasta(fapath, inpath, outpath)
		if e != nil { return h(e) }
	}

	return nil
}

func GetFastasBg(fapath string, bins []string, opre string) error {
	h := handle("GetFastas: %w")

	for _, bin := range bins {
		inpath := opre + "_" + bin + "_bg_joined.bed"
		outpath := opre + "_" + bin + "_bg_joined.fa"
		e := GetFasta(fapath, inpath, outpath)
		if e != nil { return h(e) }
	}

	return nil
}

func main() {
	bincolp := flag.Int("c", -1, "bin column")
	inpathp := flag.String("i", "", "Input path")
	oprep := flag.String("o", "", "Output prefix")
	fap := flag.String("f", "", "genome fasta file to use for generating subset fasta files")
	bgp := flag.Bool("bg", false, "Generate a background file that contains the opposite of the binned files")
	flag.Parse()
	if *bincolp == -1 { log.Fatal("missing -c") }
	if *inpathp == "" { log.Fatal("missing -i") }
	if *oprep == "" { log.Fatal("missing -o") }

	bins, e := FindBins(*bincolp, *inpathp)
	if e != nil { panic(e) }

	e = SplitByBins(*bincolp, bins, *inpathp, *oprep)
	if e != nil { panic(e) }

	e = JoinSplits(bins, *oprep)
	if e != nil { panic(e) }

	if *fap != "" {
		e = GetFastas(*fap, bins, *oprep)
		if e != nil { panic(e) }
	}

	if !*bgp {
		return
	}

	e = SplitByBinsBg(*bincolp, bins, *inpathp, *oprep)
	if e != nil { panic(e) }

	e = JoinSplitsBg(bins, *oprep)
	if e != nil { panic(e) }

	if *fap != "" {
		e = GetFastasBg(*fap, bins, *oprep)
		if e != nil { panic(e) }
	}


}
