package vsdiv

import (
	"golang.org/x/sync/errgroup"
	"github.com/jgbaldwinbrown/slide/pkg"
	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/covplots/pkg"
	"path/filepath"
	"fmt"
	"strings"
)

func RunFull() {
	args := MakeFullArgsWorkable()
	outpre := "megaseqout/out"
	if e := Full(outpre, args...); e != nil {
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
			isets[i], e = SlideAndMakeInputSet(a)
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
	return covplots.AllMultiplotParallel(cs, 0, 0, 1, true, nil)
}

func MakeInputSet(n NameSet) (covplots.InputSet, error) {
	var out covplots.InputSet
	out.Paths = append(out.Paths, n.SnpWinOutpath, n.PairInpath)
	out.Name = n.Name + "_" + n.Ref
	out.Functions = []string {
		"strip_header_some",
		"hic_pair_prop_cols_some",
		"fourcolumns",
		"combine_to_one_line_dumb",
	}
	out.FunctionArgs = []any {
		[]any{1},
		[]any{1},
		nil,
		nil,
	}
	return out, nil
}

func MakePlotArgs(outpre string, sets ...covplots.InputSet) (covplots.UltimateConfig, error) {
	var c covplots.UltimateConfig
	c.Plotfunc = "plot_self_vs_pair_pretty"
	c.PlotfuncArgs = covplots.PlotSelfVsPairArgs {
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
	c.Ylim = []float64{0, 0.25}
	c.InputSets = append(c.InputSets, sets...)
	return c, nil
}

func SlideAndMakeInputSet(n NameSet) (covplots.InputSet, error) {
	if e := SlidingGffEntryCountPaths(n.SnpInpath, n.SnpWinOutpath, float64(n.WinSize), float64(n.WinStep)); e != nil {
		return covplots.InputSet{}, e
	}
	return MakeInputSet(n)
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

	return slide.SlidingGffEntryCountFull(in, out, size, step)
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
	suffix := "_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt"
	snpinprefix := "/home/jgbaldwinbrown/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/nucdiff_all/fullset/links"
	snpinsuffix := "_snp_counts.gff.gz"
	snpoutsuf := "_snp_win.txt"
	return NameSet {
		PairInpath: nr.Name + suffix,
		SnpInpath: filepath.Join(snpinprefix, nr.Ref + snpinsuffix),
		WinSize: 100000,
		WinStep: 10000,
		SnpWinOutpath: nr.Name + snpoutsuf,
		Name: nr.Name,
		Ref: nr.Ref,
	}
}

func MakeNameSetV2(nr NameRef) NameSet {
	suffix := "_hits_100kb_dist100kb_dist100kb_mindist1kb_pairmindist1kb_named_sim800bp.txt.gz"
	snpinprefix := "/home/jgbaldwinbrown/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/nucdiff_all/fullset/links"
	snpinsuffix := "_snp_counts.gff.gz"
	snpoutsuf := "_snp_win.txt"
	return NameSet {
		PairInpath: nr.Name + "_to_" + nr.Ref + suffix,
		SnpInpath: filepath.Join(snpinprefix, nr.Ref + snpinsuffix),
		WinSize: 100000,
		WinStep: 10000,
		SnpWinOutpath: nr.Name + "_to_" + snpoutsuf,
		Name: nr.Name,
		Ref: nr.Ref,
	}
}

func MakeArgsV1(namerefs ...NameRef) []NameSet {
	as := make([]NameSet, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, MakeNameSetV1(nr))
	}
	return as
}

func MakeArgsV2(namerefs ...NameRef) []NameSet {
	as := make([]NameSet, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, MakeNameSetV2(nr))
	}
	return as
}

func MakeFullArgs() []NameSet {
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

func MakeFullArgsWorkable() []NameSet {
	return MakeArgsV1(
		NameRef{"ixw_fat", "ixw"},
		NameRef{"ixa4_sal", "ixa4"}, NameRef{"ixa4_fat", "ixa4"},
		NameRef{"ixa7_sal", "ixa7"},
		NameRef{"a7xn_sal", "ixw"},
		NameRef{"nxw_sal", "nxw"},
	)
}
