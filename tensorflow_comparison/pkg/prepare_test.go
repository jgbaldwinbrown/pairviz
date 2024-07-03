package prepfa

import (
	"fmt"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"testing"
	"strings"
	"reflect"
)

const input1 = `>1
atgtaagt`

const input2 = `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	iso1	a7	s14	w501	saw
1	1	.	A	G	0	PASS	.	GT	0/0	0/0	1/1	1/1	1/1
1	2	.	T	N	0	PASS	.	GT	0/0	0/0	1/1	1/1	1/1
1	5	.	A	C	0	PASS	.	GT	0/0	1/1	1/1	1/1	1/1`

const outputstr = `>1
atgtaagt`

var output1 = []fastats.FaEntry{fastats.FaEntry{Header: "1", Seq: "atgtaagt"}}
var output2 = []fastats.FaEntry{fastats.FaEntry{Header: "1", Seq: "atgtcagt"}}

func TestGetFa(t *testing.T) {
	fa := fastats.ParseFasta(strings.NewReader(input1))
	for f, err := range fa {
		if err != nil {
			panic(err)
		}
		fmt.Println("entry:", f)
	}
}

func TestBuildFas(t *testing.T) {
	fa, err := CollectErr(fastats.ParseFasta(strings.NewReader(input1)))
	if err != nil {
		panic(err)
	}
	fmt.Println("fa:", fa)

	_, it1, err := ReadVCF(strings.NewReader(input2))
	if err != nil {
		panic(err)
	}
	it2 := SubsetVCFCols(it1, 0, 1)
	vcf, err := CollectErr(it2)
	if err != nil {
		panic(err)
	}

	fa1, fa2, _, _ := BuildFas(fa, vcf)
	for i, _ := range fa1 {
		fa1[i].Seq = strings.ToLower(fa1[i].Seq)
	}
	for i, _ := range fa2 {
		fa2[i].Seq = strings.ToLower(fa2[i].Seq)
	}

	if !reflect.DeepEqual(fa1, output1) {
		t.Errorf("fa1 %v != output1 %v", fa1, output1)
	}
	if !reflect.DeepEqual(fa2, output2) {
		t.Errorf("fa2 %v != output2 %v", fa2, output2)
	}
}
