package prepfa

import (
	"log"
	"bufio"
	"sort"
	"fmt"
	"io"
	"flag"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/iter"
	"strings"
	"regexp"
	"github.com/jgbaldwinbrown/csvh"
)

func AppendSplit(out []string, s string, sep string) []string {
	first := ""
	found := true
	for found {
		first, s, found = strings.Cut(s, sep)
		out = append(out, first)
	}
	return out
}

func ReadGenos(line []string) ([]string, error) {
	genos := line[9:]
	ref := line[3]
	seqs := append([]string{ref}, strings.Split(line[4], ",")...)

	out := make([]string, len(genos))
	for _, geno := range genos {
		var seqi int
		_, e := fmt.Sscanf(geno, "%v", &seqi)
		if e != nil {
			return nil, e
		}
		if seqi >= len(seqs) {
			return nil, fmt.Errorf("seqi %v > len(seqs) %v; seqs: %v; geno: %v; genos: %v", seqi, len(seqs), seqs, geno, genos)
		}
		out = append(out, seqs[seqi])
	}
	return out, nil
}

func ReadWinBed(r io.Reader) iter.Iter[fastats.ChrSpan] {
	bed := fastats.ParseBed[struct{}](r, func([]string) (struct{}, error) {
		return struct{}{}, nil
	})
	return &iter.Iterator[fastats.ChrSpan]{Iteratef: func(yield func(fastats.ChrSpan) error) error {
		return bed.Iterate(func(b fastats.BedEntry[struct{}]) error {
			return yield(b.ChrSpan)
		})
	}}
}

func CollectFa(path string) ([]fastats.FaEntry, error) {
	r, err := csvh.OpenMaybeGz(path)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	return iter.Collect[fastats.FaEntry](fastats.ParseFasta(r))
}

var namesRe = regexp.MustCompile(`^#CHROM.*FORMAT\t(.*)`)

func ReadVCF(r io.Reader) (names []string, it *iter.Iterator[fastats.VcfEntry[[]string]], err error) {
	s := iter.NewScanner(r)
	for s.Scan() {
		if s.Err() != nil {
			return nil, nil, s.Err()
		}

		if match := namesRe.FindStringSubmatch(s.Text()); match != nil {
			names = strings.Split(match[1], "\t")
			break
		}
	}

	return names, &iter.Iterator[fastats.VcfEntry[[]string]]{Iteratef: func(yield func(fastats.VcfEntry[[]string]) error) error {
		var line []string
		i := 0
		h := func(i int, e error) error {
			return fmt.Errorf("ReadVCF: i: %v; e: %w", i, e)
		}
		for s.Scan() {
			var v fastats.VcfEntry[[]string]
			line = AppendSplit(line[:0], s.Text(), "\t")
			e := fastats.ParseVcfMainFields(&v, line)
			if e != nil {
				return h(i, e)
			}

			v.InfoAndSamples, e = ReadGenos(line)
			if e != nil {
				return h(i, e)
			}

			e = yield(v)
			if e != nil {
				return h(i, e)
			}
			i++
		}
		return nil
	}}, nil
}

func SubsetVCFCols(it iter.Iter[fastats.VcfEntry[[]string]], cols ...int) *iter.Iterator[fastats.VcfEntry[[]string]] {
	return &iter.Iterator[fastats.VcfEntry[[]string]]{Iteratef: func(yield func(fastats.VcfEntry[[]string]) error) error {
		return it.Iterate(func(v fastats.VcfEntry[[]string]) error {
			v2 := v
			v2.InfoAndSamples = make([]string, 0, len(cols))

			for _, col := range cols {
				if col >= len(v.InfoAndSamples) {
					return fmt.Errorf("col %v >= len(v.InfoAndSamples) %v", col, len(v.InfoAndSamples))
				}
				v2.InfoAndSamples = append(v2.InfoAndSamples, v.InfoAndSamples[col])
			}
			return yield(v2)
		})
	}}
}

func CollectVCF(path string, cols ...int) (names []string, vcf []fastats.VcfEntry[[]string], err error) {
	r, err := csvh.OpenMaybeGz(path)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	names, it1, err := ReadVCF(r)
	if err != nil {
		return nil, nil, err
	}
	it2 := SubsetVCFCols(it1, cols...)
	vcf, err = iter.Collect[fastats.VcfEntry[[]string]](it2)
	return names, vcf, err
}

func padns(buf []byte, length int) []byte {
	for len(buf) < length {
		buf = append(buf, 'N')
	}
	return buf
}

func padallns(bufs ...*[]byte) {
	if len(bufs) < 1 {
		return
	}

	longest := len(*(bufs[0]))
	for _, buf := range bufs[1:] {
		if len(*buf) > longest {
			longest = len(*buf)
		}
	}

	for _, buf := range bufs {
		*buf = padns(*buf, longest)
	}
}

func AppendString(buf []byte, s string) []byte {
	for _, b := range []byte(s) {
		buf = append(buf, b)
	}
	return buf
}

func BuildChr(entry fastats.FaEntry, vcf []fastats.VcfEntry[[]string]) (fa1, fa2 fastats.FaEntry) {
	var fa1b, fa2b strings.Builder
	var fa1buf, fa2buf []byte

	prevpos := int64(0)

	for _, v := range vcf {
		curpos := v.Start
		io.WriteString(&fa1b, entry.Seq[prevpos : curpos - 1])
		io.WriteString(&fa2b, entry.Seq[prevpos : curpos - 1])

		fa1buf = AppendString(fa1buf, v.InfoAndSamples[0])
		fa2buf = AppendString(fa2buf, v.InfoAndSamples[1])

		padallns(&fa1buf, &fa2buf)

		fa1b.Write(fa1buf)
		fa2b.Write(fa2buf)
		prevpos = curpos
	}

	fa1 = fastats.FaEntry{entry.Header, fa1b.String()}
	fa2 = fastats.FaEntry{entry.Header, fa2b.String()}

	return fa1, fa2
}

func BuildFas(fa []fastats.FaEntry, vcf []fastats.VcfEntry[[]string]) (fa1, fa2 []fastats.FaEntry) {
	chrs := make(map[string]fastats.FaEntry, len(fa))
	for _, entry := range fa {
		chrs[entry.Header] = entry
	}

	vchrs := make(map[string][]fastats.VcfEntry[[]string])
	for _, v := range vcf {
		vchrs[v.Chr] = append(vchrs[v.Chr], v)
	}
	for _, vslice := range vchrs {
		sort.Slice(vslice, func(i, j int) bool {
			return vslice[i].Start < vslice[j].Start
		})
	}

	for _, entry := range chrs {
		fa1chr, fa2chr := BuildChr(entry, vchrs[entry.Header])
		fa1 = append(fa1, fa1chr)
		fa2 = append(fa2, fa2chr)
	}

	return fa1, fa2
}

func WriteFasta(path string, it iter.Iter[fastats.FaEntry]) error {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	return it.Iterate(func(f fastats.FaEntry) error {
		_, e := fmt.Fprintf(w, ">%v\n%v\n", f.Header, f.Seq)
		return e
	})
}

func RunPrepFa() {
	fap := flag.String("fa", "", "path to .fa or .fa.gz file")
	vcfp := flag.String("vcf", "", "path to .vcf or .vcf.gz file")
	c0p := flag.Int("c0", 0, "first column")
	c1p := flag.Int("c1", 1, "second column")
	_ = flag.Int("size", 100000, "window size")
	_ = flag.Int("step", 10000, "window step")
	outprep := flag.String("o", "out", "output prefix")
	flag.Parse()

	if *fap == "" {
		panic(fmt.Errorf("mising -fa"))
	}
	if *vcfp == "" {
		panic(fmt.Errorf("missing -vcf"))
	}

	fa, err := CollectFa(*fap)
	if err != nil {
		panic(err)
	}

	log.Printf("len(fa): %v\n", len(fa))

	_, vcf, err := CollectVCF(*vcfp, *c0p, *c1p)
	if err != nil {
		panic(err)
	}

	log.Printf("len(vcf): %v\n", len(vcf))

	fa1, fa2 := BuildFas(fa, vcf)

	if err := WriteFasta((*outprep) + "_1.fa.gz", iter.SliceIter[fastats.FaEntry](fa1)); err != nil {
		panic(err)
	}
	if err := WriteFasta((*outprep) + "_2.fa.gz", iter.SliceIter[fastats.FaEntry](fa2)); err != nil {
		panic(err)
	}

	// if err = WriteWins((*outprep) + "_1.fa.gz", FaWins(fa1, *sizep, *stepp)); err != nil {
	// 	panic(err)
	// }

	// if err = WriteWins((*outprep) + "_2.fa.gz", FaWins(fa2, *sizep, *stepp)); err != nil {
	// 	panic(err)
	// }
}

// #CHROM  POS     ID      REF     ALT     QUAL    FILTER	iso1	a7	s14	w501	saw
// 2L	4541	.	A	C	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4563	.	T	A	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4573	.	A	G	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4580	.	TC	AT	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4587	.	T	C	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4595	.	AG	A	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4600	.	A	AG	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4605	.	GA	AG	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 2L	4608	.	C	G	0	PASS		GT	0/0	0/0	0/0	0/0	1/1
// 
