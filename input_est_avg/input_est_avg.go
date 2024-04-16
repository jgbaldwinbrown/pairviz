package main

import (
	"strings"
	"flag"
	"fmt"
	"encoding/json"
	"os"
	"io"
	"github.com/montanaflynn/stats"
	"github.com/sajari/regression"
)

type Ests struct {
	Name string
	ExtractBatch int
	ExtractBatchAltFpkm float64
	HicBatch int
	HicBatchAltFpkm float64
	Geno string
	Tissue string
	TissueInput float64
	EcoliInput float64
	AltFpkm float64
	PcrCycles float64

	InputEsts []float64
	InputMean float64
	HicEsts []float64
	HicMean float64
	PostSizeEsts []float64
	PostSizeMean float64
	FinalEsts []float64
	FinalMean float64

	Predictor []float64
}

func Mean(fs ...float64) float64 {
	var sum float64
	for _, f := range fs {
		sum += f
	}
	return sum / float64(len(fs))
}

func Extract(f func(Ests) float64, es ...Ests) []float64 {
	fs := make([]float64, 0, len(es))
	for _, e := range es {
		fs = append(fs, f(e))
	}
	return fs
}

func ExtractBatchMean(es []Ests) {
	sums := map[int]float64{}
	counts := map[int]float64{}
	for _, e := range es {
		counts[e.ExtractBatch]++
		sums[e.ExtractBatch] += e.AltFpkm
	}
	for i, _ := range es {
		es[i].ExtractBatchAltFpkm = sums[es[i].ExtractBatch] / counts[es[i].ExtractBatch]
	}
}

func HicBatchMean(es []Ests) {
	sums := map[int]float64{}
	counts := map[int]float64{}
	for _, e := range es {
		counts[e.HicBatch]++
		sums[e.HicBatch] += e.AltFpkm
	}
	for i, _ := range es {
		es[i].HicBatchAltFpkm = sums[es[i].HicBatch] / counts[es[i].HicBatch]
	}
}

func Filter(f func(Ests) bool, es ...Ests) []Ests {
	out := []Ests{}
	for _, e := range es {
		if f(e) {
			out = append(out, e)
		}
	}
	return out
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func MakePredictors(es ...Ests) {
	ext := map[int]int{}
	hic := map[int]int{}
	for _, e := range es {
		ext[e.ExtractBatch]++
		hic[e.HicBatch]++
	}

	for i, e := range es {
		es[i].Predictor = make([]float64, len(ext) + len(hic) + 3)
		for j, _ := range es[i].Predictor {
			es[i].Predictor[j] = -1.0
		}
		es[i].Predictor[e.ExtractBatch] = 1.0
		es[i].Predictor[e.HicBatch + len(ext)] = 1.0
		es[i].Predictor[len(ext) + len(hic)] = e.InputMean
		es[i].Predictor[len(ext) + len(hic) + 1] = e.HicMean
		es[i].Predictor[len(ext) + len(hic) + 2] = e.PostSizeMean
	}
}

func Regress(es ...Ests) *regression.Regression {
	MakePredictors(es...)

	var ds regression.DataPoints
	for _, e := range es {
		d := regression.DataPoint(e.AltFpkm, e.Predictor)
		ds = append(ds, d)
	}
	r := new(regression.Regression)
	r.SetObserved("Fpkm")

	letters := "abcdefghijklmnopqrstuvwxyz"
	for i, _ := range es[0].Predictor {
		r.SetVar(i, fmt.Sprintf("%c", letters[i]))
	}
	r.Train(ds...)
	r.Run()
	return r
}

type Flags struct {
	ToTable bool
}

func Full() {
	var f Flags
	flag.BoolVar(&f.ToTable, "t", false, "Write regression info to table for regressing in R")
	flag.Parse()

	if f.ToTable {
		FullTable(f)
	} else {
		FullRegress(f)
	}
}

func FullRegress(f Flags) {
	dec := json.NewDecoder(os.Stdin)

	var est Ests
	var ests []Ests
	for e := dec.Decode(&est); e != io.EOF; e = dec.Decode(&est) {
		if e != nil {
			panic(e)
		}
		est.InputMean = Mean(est.InputEsts...)
		est.HicMean = Mean(est.HicEsts...)
		est.PostSizeMean = Mean(est.PostSizeEsts...)
		est.FinalMean = Mean(est.FinalEsts...)
		ests = append(ests, est)
	}

	fpkms := Extract(func(e Ests) float64 { return e.AltFpkm }, ests...)
	input := Extract(func(e Ests) float64 { return e.InputMean }, ests...)
	hic := Extract(func(e Ests) float64 { return e.HicMean }, ests...)
	size := Extract(func(e Ests) float64 { return e.PostSizeMean }, ests...)
	tissue := Extract(func(e Ests) float64 { return e.TissueInput }, ests...)
	ecoli := Extract(func(e Ests) float64 { return e.EcoliInput }, ests...)

	c, e := stats.Correlation(input, fpkms)
	must(e)
	fmt.Println("input", c)

	c, e = stats.Correlation(hic, fpkms)
	must(e)
	fmt.Println("hic", c)

	c, e = stats.Correlation(size, fpkms)
	must(e)
	fmt.Println("size", c)

	c, e = stats.Correlation(tissue, fpkms)
	must(e)
	fmt.Println("tissue", c)

	c, e = stats.Correlation(ecoli, fpkms)
	must(e)
	fmt.Println("ecoli", c)

	eOverH := Extract(func(e Ests) float64 { return e.EcoliInput / e.HicMean }, ests...)
	c, e = stats.Correlation(eOverH, fpkms)
	must(e)
	fmt.Println("eOverH", c)


	eMinusH := Extract(func(e Ests) float64 { return e.EcoliInput - e.HicMean }, ests...)
	c, e = stats.Correlation(eMinusH, fpkms)
	must(e)
	fmt.Println("eMinusH", c)

	ExtractBatchMean(ests)
	HicBatchMean(ests)

	ExtractBatch := Extract(func(e Ests) float64 { return e.ExtractBatchAltFpkm }, ests...)
	c, e = stats.Correlation(ExtractBatch, fpkms)
	must(e)
	fmt.Println("ExtractBatch", c)

	HicBatch := Extract(func(e Ests) float64 { return e.HicBatchAltFpkm }, ests...)
	c, e = stats.Correlation(HicBatch, fpkms)
	must(e)
	fmt.Println("HicBatch", c)

	r := Regress(ests...)
	fmt.Println(r)
	fmt.Println(r.GetCoeffs())
	fmt.Println(r.Formula)
}

func Header() string {
	return "name\textract_batch\thic_batch\tgenotype\ttissue\ttissue_input\tecoli_input\talt_fpkm\tinput_mean\thic_mean\tpost_size_mean\tfinal_mean\tpcr_cycles\thybrid"
}

func Hybridify(geno string) string {
	switch geno {
	case "a7xn": return "hybrid"
	case "hxw": return "hybrid"
	case "ixa4": return "pure"
	case "ixa7": return "pure"
	case "ixl": return "hybrid"
	case "ixs": return "sawamura"
	case "ixw": return "hybrid"
	case "mxw": return "pure"
	case "nxw": return "pure"
	case "sxw": return "sawamura"
	default: return ""
	}
}

func OutList(e Ests) []any {
	return []any{
		e.Name,
		fmt.Sprintf(`"%v"`, e.ExtractBatch),
		fmt.Sprintf(`"%v"`, e.HicBatch),
		e.Geno,
		e.Tissue,
		e.TissueInput,
		e.EcoliInput,
		e.AltFpkm,
		e.InputMean,
		e.HicMean,
		e.PostSizeMean,
		e.FinalMean,
		e.PcrCycles,
		Hybridify(e.Geno),
	}
}

func JoinAny(as ...any) string {
	var b strings.Builder
	if len(as) > 0 {
		fmt.Fprint(&b, as[0])
	}
	for _, a := range as[1:] {
		fmt.Fprintf(&b, "\t%v", a)
	}
	return b.String()
}

func FullTable(f Flags) {
	dec := json.NewDecoder(os.Stdin)

	var est Ests
	var ests []Ests
	for e := dec.Decode(&est); e != io.EOF; e = dec.Decode(&est) {
		if e != nil {
			panic(e)
		}
		est.InputMean = Mean(est.InputEsts...)
		est.HicMean = Mean(est.HicEsts...)
		est.PostSizeMean = Mean(est.PostSizeEsts...)
		est.FinalMean = Mean(est.FinalEsts...)
		ests = append(ests, est)
	}

	fmt.Println(Header())
	for _, est := range ests {
		if _, e := fmt.Println(JoinAny(OutList(est)...)); e != nil {
			panic(e)
		}
	}
}

func main() {
	Full()
}
