package prepfa

import (
	"log"
	"bufio"
	"sort"
	"fmt"
	"io"
	"flag"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"iter"
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
	// log.Println("raw genos:", genos)
	ref := line[3]
	seqs := append([]string{ref}, strings.Split(line[4], ",")...)

	out := make([]string, 0, len(genos))
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
	// log.Println("ReadGenos out:", out)
	return out, nil
}

func ReadWinBed(r io.Reader) iter.Seq2[fastats.ChrSpan, error] {
	bed := fastats.ParseBed[struct{}](r, func([]string) (struct{}, error) {
		return struct{}{}, nil
	})
	return func(yield func(fastats.ChrSpan, error) bool) {
		for b, err := range bed {
			if ok := yield(b.ChrSpan, err); !ok {
				return
			}
		}
	}
}

func CollectFa(path string) ([]fastats.FaEntry, error) {
	r, err := csvh.OpenMaybeGz(path)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	return CollectErr(fastats.ParseFasta(r))
}

var namesRe = regexp.MustCompile(`^#CHROM.*FORMAT\t(.*)`)

func ReadVCF(r io.Reader) (names []string, it iter.Seq2[fastats.VcfEntry[[]string], error], err error) {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	for s.Scan() {
		if s.Err() != nil {
			return nil, nil, s.Err()
		}

		if match := namesRe.FindStringSubmatch(s.Text()); match != nil {
			names = strings.Split(match[1], "\t")
			break
		}
	}

	return names, func(yield func(fastats.VcfEntry[[]string], error) bool) {
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
				yield(v, h(i, e))
				return
			}

			v.InfoAndSamples, e = ReadGenos(line)
			if e != nil {
				yield(v, h(i, e))
				return
			}

			if ok := yield(v, nil); !ok {
				return
			}
			i++
		}
	}, nil
}

func SubsetVCFCols(it iter.Seq2[fastats.VcfEntry[[]string], error], cols ...int) iter.Seq2[fastats.VcfEntry[[]string], error] {
	return func(yield func(fastats.VcfEntry[[]string], error) bool) {
		for v, err := range it {
			if err != nil {
				yield(v, err)
				return
			}
			v2 := v
			v2.InfoAndSamples = make([]string, 0, len(cols))

			for _, col := range cols {
				if col >= len(v.InfoAndSamples) {
					yield(v, fmt.Errorf("col %v >= len(v.InfoAndSamples) %v", col, len(v.InfoAndSamples)))
					return
				}
				v2.InfoAndSamples = append(v2.InfoAndSamples, v.InfoAndSamples[col])
			}
			if ok := yield(v2, nil); !ok {
				return
			}
		}
	}
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
	vcf, err = CollectErr(it2)
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
	return append(buf, []byte(s)...)
}

func chrSpan(chr string, start, end int64) fastats.ChrSpan {
	return fastats.ChrSpan{Chr: chr, Span: fastats.Span{start, end}}
}

// to fix:
// single bases sometimes duplicated when unused alt allele is an n

// not printing after last vcf polymorphism

func BuildChr(entry fastats.FaEntry, vcf []fastats.VcfEntry[[]string]) (fa1, fa2 fastats.FaEntry, coord1, coord2 []CoordsPair) {
	var fa1b, fa2b strings.Builder
	var fa1buf, fa2buf []byte

	prevpos := int64(0)
	prevpos1 := int64(0)
	prevpos2 := int64(0)

	for _, v := range vcf {
		curpos := v.Start
		curpos1 := prevpos1 + (curpos - prevpos)
		curpos2 := prevpos2 + (curpos - prevpos)

		if curpos < prevpos {
			continue
		}
		io.WriteString(&fa1b, entry.Seq[prevpos : curpos])
		io.WriteString(&fa2b, entry.Seq[prevpos : curpos])

		// if prevpos <= curpos {
		// 	io.WriteString(&fa1b, entry.Seq[prevpos : curpos])
		// 	io.WriteString(&fa2b, entry.Seq[prevpos : curpos])
		// }

		coord1 = append(coord1, CoordsPair {
			chrSpan(entry.Header, prevpos, curpos),
			chrSpan(entry.Header, prevpos1, curpos1),
		})
		coord2 = append(coord2, CoordsPair {
			chrSpan(entry.Header, prevpos, curpos),
			chrSpan(entry.Header, prevpos2, curpos2),
		})

		// log.Println("fa1buf before:", fa1buf)
		// log.Println("fa2buf before:", fa2buf)

		// log.Println("v.InfoAndSamples:", v.InfoAndSamples)

		fa1buf = AppendString(fa1buf[:0], v.InfoAndSamples[0])
		fa2buf = AppendString(fa2buf[:0], v.InfoAndSamples[1])

		// log.Println("fa1buf mid:", fa1buf)
		// log.Println("fa2buf mid:", fa2buf)

		padallns(&fa1buf, &fa2buf)

		fa1b.Write(fa1buf)
		fa2b.Write(fa2buf)

		// log.Println("fa1buf after:", fa1buf)
		// log.Println("fa2buf after:", fa2buf)

		coord1 = append(coord1, CoordsPair {
			chrSpan(entry.Header, curpos, curpos + int64(len(v.Ref))),
			chrSpan(entry.Header, curpos1, curpos1 + int64(len(fa1buf))),
		})
		coord2 = append(coord2, CoordsPair {
			chrSpan(entry.Header, curpos, curpos + int64(len(v.Ref))),
			chrSpan(entry.Header, curpos2, curpos2 + int64(len(fa2buf))),
		})

		prevpos = curpos + int64(len(v.Ref))
		prevpos1 = curpos1 + int64(len(fa1buf))
		prevpos2 = curpos2 + int64(len(fa2buf))
	}

	if prevpos < int64(len(entry.Seq)) {
		io.WriteString(&fa1b, entry.Seq[prevpos : len(entry.Seq)])
		io.WriteString(&fa2b, entry.Seq[prevpos : len(entry.Seq)])

		added := int64(len(entry.Seq)) - prevpos

		coord1 = append(coord1, CoordsPair {
			chrSpan(entry.Header, prevpos, int64(len(entry.Seq))),
			chrSpan(entry.Header, prevpos1, prevpos1 + added),
		})
		coord2 = append(coord2, CoordsPair {
			chrSpan(entry.Header, prevpos, int64(len(entry.Seq))),
			chrSpan(entry.Header, prevpos2, prevpos2 + added),
		})
	}

	fa1 = fastats.FaEntry{entry.Header, fa1b.String()}
	fa2 = fastats.FaEntry{entry.Header, fa2b.String()}

	return fa1, fa2, coord1, coord2
}

type CoordsPair struct {
	Original fastats.ChrSpan
	New fastats.ChrSpan
}

func BuildFas(fa []fastats.FaEntry, vcf []fastats.VcfEntry[[]string]) (fa1, fa2 []fastats.FaEntry, coords1, coords2 []CoordsPair) {
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
		fa1chr, fa2chr, coord1, coord2 := BuildChr(entry, vchrs[entry.Header])
		fa1 = append(fa1, fa1chr)
		fa2 = append(fa2, fa2chr)
		coords1 = append(coords1, coord1...)
		coords2 = append(coords2, coord2...)
	}

	return fa1, fa2, coords1, coords2
}

func WriteFasta(path string, it iter.Seq2[fastats.FaEntry, error]) error {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for f, e := range it {
		if e != nil {
			return e
		}
		_, e = fmt.Fprintf(bw, ">%v\n%v\n", f.Header, f.Seq)
		if e != nil {
			return e
		}
	}
	return nil
}

func WriteCoords(path string, it iter.Seq2[CoordsPair, error]) error {
	w, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for c, e := range it {
		if e != nil {
			return e
		}
		_, e = fmt.Fprintf(bw, "%v\t%v\t%v\t%v\t%v\t%v\n",
			c.Original.Chr, c.Original.Start, c.Original.End,
			c.New.Chr, c.New.Start, c.New.End,
		)
		if e != nil {
			return e
		}
	}
	return nil
}

func coordConv(pos int, coords []CoordsPair) int {
	i := sort.Search(len(coords), func(i int) bool { return coords[i].Original.End > int64(pos) })
	return pos + int(coords[i].New.Start - coords[i].Original.Start)
}

func MakeWins(fa []fastats.FaEntry, coords []CoordsPair, size int, step int, width int) iter.Seq2[fastats.ChrSpan, error] {
	return func(yield func(fastats.ChrSpan, error) bool) {
		chrcoords := make(map[string][]CoordsPair, len(fa))
		for _, pair := range coords {
			chrcoords[pair.Original.Chr] = append(chrcoords[pair.Original.Chr], pair)
		}

		half := size / 2
		halfwidth := width / 2
		for _, entry := range fa {
			for mid := size / 2; mid + half < len(entry.Seq); mid += step {
				midconv := coordConv(mid, chrcoords[entry.Header])
				span := chrSpan(entry.Header, int64(midconv - halfwidth), int64(midconv + halfwidth))
				if ok := yield(span, nil); !ok {
					return
				}
			}
		}
	}
}

func FaFixedWins(fa []fastats.FaEntry, wins iter.Seq2[fastats.ChrSpan, error]) iter.Seq2[fastats.FaEntry, error] {
	return func(yield func(fastats.FaEntry, error) bool) {
		chrs := make(map[string]fastats.FaEntry, len(fa))
		for _, entry := range fa {
			chrs[entry.Header] = entry
		}

		for s, err := range wins {
			if err != nil {
				yield(fastats.FaEntry{}, err)
				return
			}
			if s.Span.Start < 0 || s.Span.End > int64(len(chrs[s.Chr].Seq)) {
				log.Printf("Trying to extract win out of bounds; chr: %v; chrlen: %v; span: %v\n", s.Chr, len(chrs[s.Chr].Seq), s.Span)
				continue
			}
			out, err := fastats.ExtractOne(chrs[s.Chr], s.Span)
			if err != nil {
				yield(fastats.FaEntry{}, err)
				return
			}
			if ok := yield(out, nil); !ok {
				return
			}
		}
	}
}

type PrepFaFlags struct {
	Fa string
	Vcf string
	C0 int
	C1 int
	Size int
	Step int
	Width int
	Outpre string
}

func GetPrepFaFlags() (PrepFaFlags, error) {
	var f PrepFaFlags

	flag.StringVar(&f.Fa, "fa", "", "path to .fa or .fa.gz file")
	flag.StringVar(&f.Vcf, "vcf", "", "path to .vcf or .vcf.gz file")
	flag.IntVar(&f.C0, "c0", 0, "first column")
	flag.IntVar(&f.C1, "c1", 1, "second column")
	flag.IntVar(&f.Size, "size", 100000, "window size")
	flag.IntVar(&f.Step, "step", 10000, "window step")
	flag.IntVar(&f.Width, "width", 90000, "portion of center of window to output")
	flag.StringVar(&f.Outpre, "o", "out", "output prefix")

	flag.Parse()

	if f.Fa == "" {
		panic(fmt.Errorf("mising -fa"))
	}
	if f.Vcf == "" {
		panic(fmt.Errorf("missing -vcf"))
	}

	return f, nil
}

func PrepFa(f PrepFaFlags) error {
	fa, err := CollectFa(f.Fa)
	if err != nil {
		panic(err)
	}

	// log.Printf("len(fa): %v\n", len(fa))

	_, vcf, err := CollectVCF(f.Vcf, f.C0, f.C1)
	if err != nil {
		panic(err)
	}

	// log.Printf("len(vcf): %v\n", len(vcf))

	fa1, fa2, coords1, coords2 := BuildFas(fa, vcf)

	if err := WriteFasta((f.Outpre) + "_1.fa.gz", AddErr(SliceIter(fa1))); err != nil {
		panic(err)
	}
	if err := WriteFasta((f.Outpre) + "_2.fa.gz", AddErr(SliceIter(fa2))); err != nil {
		panic(err)
	}
	if err := WriteCoords((f.Outpre) + "_1_coords.bed.gz", AddErr(SliceIter(coords1))); err != nil {
		panic(err)
	}
	if err := WriteCoords((f.Outpre) + "_2_coords.bed.gz", AddErr(SliceIter(coords2))); err != nil {
		panic(err)
	}

	wins1 := MakeWins(fa, coords1, f.Size, f.Step, f.Width)
	wins2 := MakeWins(fa, coords2, f.Size, f.Step, f.Width)

	if err = WriteFasta((f.Outpre) + "_wins_1.fa.gz", FaFixedWins(fa1, wins1)); err != nil {
		panic(err)
	}

	if err = WriteFasta((f.Outpre) + "_wins_2.fa.gz", FaFixedWins(fa2, wins2)); err != nil {
		panic(err)
	}

	return nil
}

func RunPrepFa() {
	flags, e := GetPrepFaFlags()
	if e != nil {
		panic(e)
	}

	e = PrepFa(flags)
	if e != nil {
		panic(e)
	}
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
