package register

import (
	"encoding/json"
	"io"
	"os"
	"fmt"
	"strings"
)

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

func MakeArgs(maxdist int64, namerefs ...NameRef) []Job {
	as := make([]Job, 0, len(namerefs))
	for _, nr := range namerefs {
		as = append(as, Job{
			Inpath: fmt.Sprintf("../rehead/%v_to_%v_rehead.pairs.gz", nr.Name, nr.Ref),
			Outpath: fmt.Sprintf("%v_to_%v_rehead_registers_max%v.txt", nr.Name, nr.Ref, BpFormat(maxdist)),
			Plot: true,
			Plotoutpath: fmt.Sprintf("%v_to_%v_rehead_registers_max%v.pdf", nr.Name, nr.Ref, BpFormat(maxdist)),
			Plotname: Pnames(nr.Name),
			Mindist: 0,
			Maxdist: maxdist,
		})
	}
	return as
}

func MakeVerySmallArgs() []Job {
	return MakeArgs(1000, NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"})
}

func MakeMinimalArgs() []Job {
	return MakeArgs(1000, NameRef{"nxw_adult_mini", "nxw"})
}

func MakeMicroArgs() []Job {
	return MakeArgs(1000, NameRef{"nxw_adult_micro", "nxw"}, NameRef{"nxw_sal_micro", "nxw"})
}

func MakeFullArgs() []Job {
	return MakeArgs(1000,
		NameRef{"ixw_sal", "ixw"}, NameRef{"ixw_adult", "ixw"}, NameRef{"ixw_brain", "ixw"}, NameRef{"ixw_fat", "ixw"},
		NameRef{"ixa4_sal", "ixa4"}, NameRef{"ixa4_adult", "ixa4"}, NameRef{"ixa4_brain", "ixa4"}, NameRef{"ixa4_fat", "ixa4"},
		NameRef{"ixa7_sal", "ixa7"}, NameRef{"ixa7_adult", "ixa7"},
		NameRef{"a7xn_sal", "ixw"}, NameRef{"a7xn_adult", "ixw"},
		NameRef{"nxw_sal", "nxw"}, NameRef{"nxw_adult", "nxw"},
		NameRef{"mxw_sal", "mxw"}, NameRef{"mxw_adult", "mxw"},
		NameRef{"hxw_sal", "ixw"}, NameRef{"ixl_sal", "ixw"},
		NameRef{"ixs_sal", "ixs"}, NameRef{"sxw_sal", "sxw"},
	)
}

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

func FullPrintJobs() {
	if e := PrintAllJobs(os.Stdout, MakeMicroArgs()); e != nil {
		panic(e)
	}
}

