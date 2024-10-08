package register

import (
	"flag"
	"encoding/json"
	"io"
	"os"
	"fmt"
	"strings"
)

// Format a number of basepairs into a human-readable format
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

// Name here is the name of a genotype, and Ref is the name of the reference genome used for alignment
type NameRef struct {
	Name string
	Ref string
}

// Convert a name of the format "genotype_tissue" into pretty-printed genotype and tissue, space separated
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

// Generate jobs to run starting from namerefs
func MakeArgs(maxdist int64, namerefs ...NameRef) []Job {
	as := make([]Job, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, Job{
			Inpath: fmt.Sprintf("../rehead/%v_to_%v_rehead.pairs.gz", nr.Name, nr.Ref),
			Outpath: fmt.Sprintf("%v_to_%v_rehead_registers_max%v.txt", nr.Name, nr.Ref, BpFormat(maxdist)),
			Plot: true,
			Plotoutpath: fmt.Sprintf("%v_to_%v_rehead_registers_max%v.pdf", nr.Name, nr.Ref, BpFormat(maxdist)),
			DirPlotoutpath: fmt.Sprintf("%v_to_%v_rehead_registers_max%v_dir.pdf", nr.Name, nr.Ref, BpFormat(maxdist)),
			Plotname: Pnames(nr.Name),
			Mindist: 0,
			Maxdist: maxdist,
		})
	}
	return as
}

// Generate jobs to run in the "final" folder
func MakeArgsFinal(maxdist int64, namerefs ...NameRef) []Job {
	as := make([]Job, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, Job{
			Register: false,
			Inpath: fmt.Sprintf("../../rehead/%v_to_%v_rehead.pairs.gz", nr.Name, nr.Ref),
			Outpath: fmt.Sprintf("%v_to_%v/%v_to_%v_rehead_registers_max%v.txt", nr.Name, nr.Ref, nr.Name, nr.Ref, BpFormat(maxdist)),
			Plot: true,
			Plotoutpath: fmt.Sprintf("%v_to_%v/%v_to_%v_rehead_registers_max%v.pdf", nr.Name, nr.Ref, nr.Name, nr.Ref, BpFormat(maxdist)),
			DirPlotoutpath: fmt.Sprintf("%v_to_%v/%v_to_%v_rehead_registers_max%v_dir.pdf", nr.Name, nr.Ref, nr.Name, nr.Ref, BpFormat(maxdist)),
			Plotname: Pnames(nr.Name),
			Mindist: 0,
			Maxdist: maxdist,
		})
	}
	return as
}

func MakeVerySmallArgs() []Job {
	return MakeArgs(10000, NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"})
}

func MakeMinimalArgs() []Job {
	return MakeArgs(10000, NameRef{"nxw_adult_mini", "nxw"})
}

func MakeMicroArgs() []Job {
	return MakeArgs(10000, NameRef{"nxw_adult_micro", "nxw"}, NameRef{"nxw_sal_micro", "nxw"})
}

func MakeFullNameRefs() []NameRef {
	return []NameRef {
		NameRef{"ixw_sal", "ixw"}, NameRef{"ixw_adult", "ixw"}, NameRef{"ixw_brain", "ixw"}, NameRef{"ixw_fat", "ixw"},
		NameRef{"ixa4_sal", "ixa4"}, NameRef{"ixa4_adult", "ixa4"}, NameRef{"ixa4_brain", "ixa4"}, NameRef{"ixa4_fat", "ixa4"},
		NameRef{"ixa7_sal", "ixa7"}, NameRef{"ixa7_adult", "ixa7"},
		NameRef{"a7xn_sal", "a7xn"}, NameRef{"a7xn_adult", "a7xn"},
		NameRef{"a7xn_sal", "ixw"}, NameRef{"a7xn_adult", "ixw"},
		NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"},
		NameRef{"mxw_sal", "mxw"}, NameRef{"mxw_adult", "mxw"},
		NameRef{"hxw_sal", "ixw"}, NameRef{"ixl_sal", "ixw"},
		NameRef{"ixs_sal", "ixs"}, NameRef{"sxw_sal", "sxw"},
	}
}

func MakeFullArgs() []Job {
	return MakeArgs(3000, MakeFullNameRefs()...)
}

func MakeFullArgsFinal() []Job {
	return MakeArgsFinal(3000, MakeFullNameRefs()...)
}

// Write all of the jobs as JSON for passing to the register program
func PrintAllJobs(w io.Writer, jobs []Job) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, j := range jobs {
		e := enc.Encode(j)
		if e != nil {
			return e
		}
	}
	return nil
}

type Flags struct {
	Final bool
}

func FullPrintJobs() {
	var f Flags
	flag.BoolVar(&f.Final, "f", false, "Final version")
	flag.Parse()

	var args []Job
	if f.Final {
		args = MakeFullArgsFinal()
	} else {
		args = MakeFullArgs()
	}

	if e := PrintAllJobs(os.Stdout, args); e != nil {
		panic(e)
	}
}

// ../../rehead/ixw_brain_to_ixw_rehead.pairs.gz
