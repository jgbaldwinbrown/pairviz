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

func SplitBin(col int, bin string, path, opath string) error {
	h := handle("SplitBin: %w")

	r, cr, e := OpenCsv(path)
	if e != nil { return h(e) }
	defer r.Close()

	w, cw, e := CreateCsv(opath)
	if e != nil { return h(e) }
	defer w.Close()
	defer cw.Flush()

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) <= col {
			continue
		}

		if line[col] == bin {
			e = cw.Write(line)
			if e != nil { return h(e) }
		}
	}

	return nil
}

func SplitByBins(col int, bins []string, path, opre string) error {
	h := handle("SplitByBins: %w")

	for _, bin := range bins {
		opath := opre + bin + ".bed"
		e := SplitBin(col, bin, path, opath)
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

	e = cmd.Run()
	if e != nil { return h(e) }

	return nil
}

func JoinSplits(bins []string, opre string) error {
	for _, bin := range bins {
		inpath := opre + bin + ".bed"
		outpath := opre + bin + "_joined.bed"
		e := JoinSplit(inpath, outpath)
		if e != nil { return fmt.Errorf("JoinSplits: %w") }
	}
	return nil
}

func main() {
	bincolp := flag.Int("c", -1, "bin column")
	inpathp := flag.String("i", "", "Input path")
	oprep := flag.String("o", "", "Output prefix")
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
}

/*
 78 func MergeHitsString() string {
 79         return `#!/bin/bash
 80 set -e
 81 
 82 mawk -F "\t" -v OFS="\t" '$'${1}' >= '${2}' || $'${1}' == "inf"{
 83         $3=sprintf("%d", $3);
 84         if ($2 < 0) { $2 = 0 };
 85         if ($3 < 0) { $3 = 0 };
 86         print $0
 87 }' \
 88 > ${3}_thresholded.bed
 89 
 90 bedtools merge -i ${3}_thresholded.bed > ${3}_thresh_merge.bed`
 91 }
 92 

*/
