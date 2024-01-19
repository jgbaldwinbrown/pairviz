package pairviz

import (
	"github.com/sajari/regression"
	"github.com/jgbaldwinbrown/iter"
	"bufio"
	"flag"
	"encoding/json"
	"os"
	"fmt"
	"io"
	"github.com/jgbaldwinbrown/csvh"
)

type BatchInfo struct {
	Name string
	Batch int
}

type BatchInfoTable struct {
	Infos []BatchInfo
	NameToBatch map[string]int
}

func GetBatchInfo(r io.Reader) (BatchInfoTable, error) {
	cr := csvh.CsvIn(r)
	var bit BatchInfoTable
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return bit, e
		}
		if len(l) < 2 {
			return bit, fmt.Errorf("GetBatchInfo: line %v too short", l)
		}

		var bi BatchInfo
		_, e = csvh.Scan(l, &bi.Name, &bi.Batch)
		if e != nil {
			return bit, e
		}
		bit.Infos = append(bit.Infos, bi)
	}

	bit.NameToBatch = map[string]int{}
	for _, bi := range bit.Infos {
		bit.NameToBatch[bi.Name] = bi.Batch
	}

	return bit, nil
}

func GetBatchInfoPath(path string) (BatchInfoTable, error) {
	r, e := OpenMaybeGz(path)
	if e != nil {
		return BatchInfoTable{}, e
	}
	defer r.Close()

	return GetBatchInfo(r)
}

func MakeControlLine(y float64, name string, t BatchInfoTable) []float64 {
	x := make([]float64, len(t.Infos) + 1)
	x[0] = y
	x[t.NameToBatch[name] + 1] = 1
	return x
}

func GetBatchControlStatTable(controlChr string, t BatchInfoTable, it iter.Iter[JsonOutStat]) ([][]float64, error) {
	var out [][]float64
	err := it.Iterate(func(j JsonOutStat) error {
		if j.Chr == controlChr {
			out = append(out, MakeControlLine(float64(j.AltFpkm), j.Name, t))
		}
		return nil
	})
	return out, err
}

func TrainTable(table [][]float64, depcol int, depname string, indepcols []int, indepnames []string) *regression.Regression {
	totrain := make(regression.DataPoints, len(table))
	for i, row := range table {
		indeps := make([]float64, len(indepcols))
		for j, col := range indepcols {
			indeps[j] = row[col]
		}
		totrain[i] = regression.DataPoint(row[depcol], indeps)
	}

	r := new(regression.Regression)
	r.SetObserved(depname)
	for i, name := range indepnames {
		r.SetVar(i, name)
	}
	r.Train(totrain...)
	r.Run()
	return r
}

func BuildModel(table [][]float64) *regression.Regression {
	indepcols := make([]int, len(table[0]) - 1)
	indepnames := make([]string, len(table[0]) - 1)
	for i := 0; i < len(indepcols); i++ {
		indepcols[i] = i + 1
		indepnames[i] = fmt.Sprint(i + 1)
	}

	return TrainTable(
		table,
		0,
		"AltFpkm",
		indepcols,
		indepnames,
	)
}

func Residual(y float64, x []float64, model *regression.Regression) (float64, error) {
	p, e := model.Predict(x)
	if e != nil {
		return 0, e
	}
	return y - p, nil
}

func ResidualFromAltFpkm(x JsonOutStat, t BatchInfoTable, model *regression.Regression) JsonOutStat {
	line := MakeControlLine(float64(x.TargetFpkm), x.Name, t)
	targetFpkm, err := Residual(line[0], line[1:], model)
	if err != nil {
		panic(err)
	}
	x.TargetFpkm = JsonFloat(targetFpkm)
	return x
}

func ResidualFromAltFpkmAll(it iter.Iter[JsonOutStat], t BatchInfoTable, model *regression.Regression) *iter.Iterator[JsonOutStat] {
	return &iter.Iterator[JsonOutStat]{Iteratef: func(yield func(JsonOutStat) error) error {
		return it.Iterate(func(x JsonOutStat) error {
			j := ResidualFromAltFpkm(x, t, model)
			return yield(j)
		})
	}}
}

// WriteModel(*outmeanp, batchInfo, controlTable, model)

func WriteModel(path string, bit BatchInfoTable, controls [][]float64, model *regression.Regression) error {
	r, e := csvh.CreateMaybeGz(path)
	if e != nil {
		return e
	}

	enc := json.NewEncoder(r)
	enc.SetIndent("", "\t")

	if e = enc.Encode(bit); e != nil {
		return e
	}
	if e = enc.Encode(controls); e != nil {
		return e
	}
	if e = enc.Encode(model); e != nil {
		return e
	}

	return nil
}

// func WriteModel(path string, control, exp JsonOutStat) error {
// 	h := func(e error) error {
// 		return fmt.Errorf("WriteMeansPath: %w", e)
// 	}
// 
// 	w, e := os.Create(path)
// 	if e != nil {
// 		return h(e)
// 	}
// 	defer w.Close()
// 
// 	enc := json.NewEncoder(w)
// 
// 	if e = enc.Encode(control); e != nil {
// 		return h(e)
// 	}
// 	if e = enc.Encode(exp); e != nil {
// 		return h(e)
// 	}
// 	return nil
// }

func FullEcnormLm() {
	batchInfoPathp := flag.String("batch", "", "path to batch info tab-delimited file (name, batch_name)")
	controlChr := flag.String("c", "", "Chromosome to use as control (required)")
	inpath := flag.String("i", "", "Input path (default stdin)")
	outmeanp := flag.String("mo", "", "Path to output means to (default discard)")

	flag.Parse()
	if *controlChr == "" {
		panic(fmt.Errorf("missing -c"))
	}
	if *inpath == "" {
		panic(fmt.Errorf("missing -i"))
	}
	if *batchInfoPathp == "" {
		panic(fmt.Errorf("missing -batch"))
	}

	batchInfo, e := GetBatchInfoPath(*batchInfoPathp)

	r, e := OpenMaybeGz(*inpath)
	Must(e)

	it := ParsePairvizOut(r)
	controlTable, e := GetBatchControlStatTable(*controlChr, batchInfo, it)
	Must(e)
	model := BuildModel(controlTable)
	Must(r.Close())
	Must(e)

	if *outmeanp != "" {
		WriteModel(*outmeanp, batchInfo, controlTable, model)
	}

	r, e = OpenMaybeGz(*inpath)
	Must(e)
	defer func() { Must(r.Close()) }()

	it = ParsePairvizOut(r)
	transit := ResidualFromAltFpkmAll(it, batchInfo, model)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	enc := json.NewEncoder(w)
	err := transit.Iterate(func(j JsonOutStat) error {
		return enc.Encode(j)
	})
	Must(err)
}
