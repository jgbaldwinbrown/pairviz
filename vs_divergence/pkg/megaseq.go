package vsdiv

import (
	"golang.org/x/sync/errgroup"
	"github.com/jgbaldwinbrown/slide/pkg"
	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/covplots/pkg"
	"flag"
	"io"
	"os"
	"fmt"
	"strings"
)

func main() {
	RunFull()
}

func RunFull() {
	args := MakeFullArgsMinus2()
	outpre := "megaseqout/out"
	if e := Full((outpre, args...); e != nil {
		panic(e)
	}
}

func Full(outpre string, args ...NameSet) error {
	var g errgroup.Group
	isets := make([]covplots.InputSet, len(args))
	for i, a := range args {
		i := i
		a := a
		g.Go(func() error {
			var e error
			isets[i], e = SlideAndMakeInputSet(arg)
			return e
		})
	}
	if e := g.Wait(); e != nil {
		return e
	}
	c, e := MakePlotArgs(outpre, isets...)
	if e != nil {
		return e
	}
	return Plot(c)
}

func Plot(cs ...covplots.UltimateConfig) error {
	return AllMultiplotParallel(cs, 0, 0, 1, true, nil)
}

func MakeInputSet(n NameSet) (covplots.InputSet, error) {
	var out covplots.InputSet
	out.Paths = append(out.Paths, n.SnpWinOutpath, n.PairInpath)
	out.Name = s.Name + "_" + s.Ref
	out.Functions = []string {
		"strip_header_some",
		"hic_pair_prop_cols_some",
		"fourcolumns",
		"combine_to_one_line_dumb",
	}
	out.FunctionArgs = []any {
		[]int{1},
		[]int{1]},
		nil,
		nil,
	}
	return out
}

func MakePlotArgs(outpre string, sets ...covplots.InputSet) (covplots.UltimateConfig, error) {
	var c covplots.UltimateConfig
	c.Plotfunc = "plot_self_vs_pair_pretty",
	c.Plotfuncargs = covplots.PlotSelfVsPairArgs {
		Xmin: 0.0,
		Xmax: 8000.0,
		Ylab: "Pairing rate",
		Xlab: "SNPs per 100kb",
		Width: 8,
		Height: 6,
		ResScale: 300,
		TextSize: 18,
	}
	c.Fullchr = true
	c.Outpre = outpre
	c.Ylim = []int{0, 0.25}
	c.InputSets = append(c.InputSets, sets...)
	return c, nil
}

func SlideAndMakeInputSet(set NameSet) (covplots.InputSet, error) {
	if e := SlidingGffEntryCountPaths(SnpInpath, SnpWinOutpath, WinSize, WinStep); e != nil {
		return covplots.InputSet{}, e
	}
	return MakeInputSet(set)
}

func SlidingGffEntryCountPaths(src, dst string, size float64, step float64) (err error) {
	in, e := csvh.OpenMaybeGz(src)
	if e != nil {
		return e
	}
	defer in.Close()

	out, e := csvh.CreateMaybeGz(dst)
	if e != nil {
		return e
	}
	defer func() { csvh.DeferE(&err, out.Close()) }()

	return SlidingGffEntryCountFull(in, out, size, step)
}

func BpFormat(bp int64) string {
	if bp >= 1e15 && bp % 1e15 == 0 {
		return fmt.Sprintf("%vEb", bp / 1e15)
	}
	if bp >= 1e12 && bp % 1e12 == 0 {
		return fmt.Sprintf("%vPb", bp / 1e12)
	}
	if bp >= 1e9 && bp % 1e9 == 0 {
		return fmt.Sprintf("%vGb", bp / 1e9)
	}
	if bp >= 1e6 && bp % 1e6 == 0 {
		return fmt.Sprintf("%vMb", bp / 1e6)
	}
	if bp >= 1000 && bp % 1000 == 0 {
		return fmt.Sprintf("%vkb", bp / 1000)
	}

	return fmt.Sprintf("%vbp", bp)
}

type NameSet struct {
	PairInpath string
	SnpInpath string
	WinSize int
	WinStep int
	SnpWinOutpath string
	Name string
	Ref string
}

type NameRef struct {
	Name string
	Ref string
	Outpre string
}

func Pnames(name string) string {
	pnames := map[string]string {
		"ixa4": "Iso1 X A4",
		"ixa7": "Iso1 X A7",
		"a7xn": "A7 X Nueva",
		"ixw": "Iso1 X w501",
		"nxw": "Nueva X w501",
		"mxw": "M252 X w501",
		"ixs": "Iso1 X Sawamura",
		"sxw": "Sawamura X w501",
		"hxw": "Hmr X w501",
		"ixl": "Hmr X w501",
	}

	tissue := map[string]string {
		"sal": "larval salivary gland",
		"adult": "adult head & thorax",
		"brain": "larval brain & imaginal discs",
		"fat": "larval fat body",
	}

	fields := strings.Split(name, "_")
	return fmt.Sprintf("%v %v", pnames[fields[0]], tissue[fields[1]])
}

func MakeNameSetV1(nr NameRef) NameSet {
	suffix := "_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz"
	snpoutsuf := "_snp_win.txt"
	return NameSet {
		PairInpath: nr.Name + suffix,
		SnpInpath string // in progress
		WinSize: 100000,
		WinStep: 10000,
		SnpWinOutpath: nr.Name + snpoutsuf,
		Name: nr.Name,
		Ref: nr.Ref,
	}
}

func MakeNameSetV1(nr NameRef) NameSet {
	suffix := "_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz"
	snpoutsuf := "_snp_win.txt"
	return NameSet {
		PairInpath: nr.Name + "_to_" + nr.Ref + suffix,
		SnpInpath string // in progress
		WinSize: 100000,
		WinStep: 10000,
		SnpWinOutpath: nr.Name + "_to_" + snpoutsuf,
		Name: nr.Name,
		Ref: nr.Ref,
	}
}

func MakeArgsV1(namerefs ...NameRef) []NameSet {
	as := make([]string, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, MakeNameSetV1(nr)))
	}
	return as
}

func MakeArgsV2(namerefs ...NameRef) []NameSet {
	suffix := "_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz"
	as := make([]string, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, nr.Name + "_to_" + nr.Ref + suffix))
	}
	return as
}

func MakeFullArgs() []string {
	return MakeArgsV2(
		NameRef{"ixw_sal", "ixw"}, NameRef{"ixw_adult", "ixw"}, NameRef{"ixw_brain", "ixw"}, NameRef{"ixw_fat", "ixw"},
		NameRef{"ixa4_sal", "ixa4"}, NameRef{"ixa4_adult", "ixa4"}, NameRef{"ixa4_brain", "ixa4"}, NameRef{"ixa4_fat", "ixa4"},
		NameRef{"ixa7_sal", "ixa7"}, NameRef{"ixa7_adult", "ixa7"},
		NameRef{"a7xn_sal", "ixw"}, NameRef{"a7xn_adult", "ixw"},
		NameRef{"a7xn_sal", "a7xn"}, NameRef{"a7xn_adult", "a7xn"},
		NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"},
		NameRef{"mxw_sal", "mxw"}, NameRef{"mxw_adult", "mxw"},
		NameRef{"hxw_sal", "ixw"}, NameRef{"ixl_sal", "ixw"},
		NameRef{"ixs_sal", "ixs"}, NameRef{"sxw_sal", "sxw"},
	)
}

func MakeFullArgsMinus2() []NameSet {
	return MakeArgsV1(
		NameRef{"ixw_adult", "ixw"}, NameRef{"ixw_brain", "ixw"}, NameRef{"ixw_fat", "ixw"},
		NameRef{"ixa4_sal", "ixa4"}, NameRef{"ixa4_adult", "ixa4"}, NameRef{"ixa4_brain", "ixa4"}, NameRef{"ixa4_fat", "ixa4"},
		NameRef{"ixa7_sal", "ixa7"}, NameRef{"ixa7_adult", "ixa7"},
		NameRef{"a7xn_sal", "ixw"}, NameRef{"a7xn_adult", "ixw"},
		NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"},
		NameRef{"mxw_sal", "mxw"},
		NameRef{"hxw_sal", "ixw"}, NameRef{"ixl_sal", "ixw"},
		NameRef{"ixs_sal", "ixs"}, NameRef{"sxw_sal", "sxw"},
	)
}

// ./a7xn_adult_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_adult_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_adult_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_adult_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_adult_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./hxw_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./hxw_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./hxw_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./hxw_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./hxw_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_adult_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_adult_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_adult_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_adult_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_adult_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_brain_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_brain_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_brain_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_brain_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_brain_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_fat_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_fat_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_fat_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_fat_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_fat_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa4_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_adult_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_adult_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_adult_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_adult_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_adult_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixa7_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixl_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixl_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixl_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixl_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixl_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixs_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixs_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixs_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixs_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_adult_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_adult_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_adult_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_adult_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_adult_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_brain_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_brain_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_brain_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_brain_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_brain_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_fat_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_fat_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_fat_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_fat_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./ixw_fat_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./mxw_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./mxw_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./mxw_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./mxw_sal_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./mxw_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_adult_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_adult_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_adult_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_adult_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./nxw_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./sxw_sal_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./sxw_sal_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./sxw_sal_hits_10Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./sxw_sal_hits_1Mb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt
// ./a7xn_sal_to_a7xn_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./a7xn_sal_to_a7xn_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./a7xn_sal_to_a7xn_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./a7xn_sal_to_ixw_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./a7xn_sal_to_ixw_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./a7xn_sal_to_ixw_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./hxw_sal_to_ixw_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./hxw_sal_to_ixw_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixl_sal_to_ixw_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixl_sal_to_ixw_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixl_sal_to_ixw_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixw_brain_to_ixw_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixw_brain_to_ixw_hits_10kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./ixw_brain_to_ixw_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./mxw_sal_to_mxw_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz
// ./mxw_sal_to_mxw_hits_1kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz


// pair_vs_structdiff_centro.json
// 
// slide_gff_entry_count -s 100000 -t 10000 <out_ref_snps.gff > snp_counts.bed
// slide_gff_bp_covered -s 100000 -t 10000 <out_ref_struct.gff > struct_bp_covered.bed
// 
// slide_gff_entry_count -s 100000 -t 10000 <ixa_ref_snps.gff > ixa_snp_counts.bed
// slide_gff_bp_covered -s 100000 -t 10000 <ixa_ref_struct.gff > ixa_struct_bp_covered.bed
// 
// slide_gff_entry_count -s 100000 -t 10000 <wxw_ref_snps.gff > wxw_snp_counts.bed
// slide_gff_bp_covered -s 100000 -t 10000 <wxw_ref_struct.gff > wxw_struct_bp_covered.bed
