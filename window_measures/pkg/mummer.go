package windif

import (
	"regexp"
	"io"
	"os/exec"
	"os"
	"path/filepath"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/csvh"
)

func Mummer(sub, quer []fastats.FaEntry) (cmd *exec.Cmd, del func(), err error) {
	dir, e := os.MkdirTemp(".", "blast_*_dir")
	if e != nil {
		return nil, nil, e
	}

	spath := filepath.Join(dir, "sub.fa")
	if e := WriteFaPath(spath, sub); e != nil {
		return nil, nil, e
	}
	qpath := filepath.Join(dir, "quer.fa")
	if e := WriteFaPath(qpath, quer); e != nil {
		return nil, nil, e
	}

	cmd = exec.Command("mummer", "-b", spath, qpath)
	// cmd := exec.Command("nucmer", "-l", "100", "-prefix", npre, spath, qpath)
	return cmd, func(){os.RemoveAll(dir)}, nil
}

type MummerMatch struct {
	QChr string
	RStart int64
	QStart int64
	Length int64
}

var headRe = regexp.MustCompile(`^> (.*)`)

func AppendMummerMatches(r io.Reader, matches []MummerMatch) ([]MummerMatch, error) {
	chr := ""
	cr := csvh.CsvIn(r)
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return nil, e
		}
		if len(l) > 0 {
			head := headRe.FindStringSubmatch(l[0])
			if head != nil {
				chr = head[1]
				continue
			}
		}

		var m MummerMatch
		m.QChr = chr
		_, e := csvh.Scan(l, &m.RStart, &m.QStart, &m.Length)
		if e != nil {
			return nil, e
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func MummerMatchBp(wp Winpair) (int64, error) {
	ec := make(chan error)
	bl, del, e := Mummer(
		[]fastats.FaEntry{wp.Fa1},
		[]fastats.FaEntry{wp.Fa2},
	)
	if e != nil {
		return 0, e
	}
	defer del()

	bl.Stderr = os.Stderr
	pr, e := bl.StdoutPipe()
	if e != nil {
		return 0, e
	}

	go func() {
		ec <- bl.Run()
	}()

	var stats []MummerMatch
	stats, e = AppendMummerMatches(pr, stats)
	if e != nil {
		return 0, e
	}
	var sum int64
	for _, stat := range stats {
		sum += stat.Length
	}
	return sum, nil
}
