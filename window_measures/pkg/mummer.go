package windif

import (
	"strings"
	"fmt"
	"bufio"
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
		return nil, nil, fmt.Errorf("Mummer: os.MkdirTemp: %w", e)
	}

	spath := filepath.Join(dir, "sub.fa")
	if e := WriteFaPath(spath, sub); e != nil {
		return nil, nil, fmt.Errorf("Mummer: WriteFaPath: %w", e)
	}
	qpath := filepath.Join(dir, "quer.fa")
	if e := WriteFaPath(qpath, quer); e != nil {
		return nil, nil, fmt.Errorf("Mummer: WriteFaPath2: %w", e)
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
var whiteRe = regexp.MustCompile(`[ \t\n]+`)

func AppendMummerMatches(r io.Reader, matches []MummerMatch) ([]MummerMatch, error) {
	chr := ""
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	for s.Scan() {
		if s.Err() != nil {
			return nil, fmt.Errorf("AppendMummerMatches s.Err(): %w", s.Err())
		}
		if head := headRe.FindStringSubmatch(s.Text()); head != nil {
			chr = head[1]
			continue
		}

		fields := whiteRe.Split(s.Text(), -1)
		var m MummerMatch
		m.QChr = chr
		_, e := csvh.Scan(fields[1:], &m.RStart, &m.QStart, &m.Length)
		if e != nil {
			return nil, fmt.Errorf("AppendMummerMatches: csvh.Scan: %w", e)
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func MummerMatchBp(wp Winpair) (int64, error) {
	h := func(e error) (int64, error) {
		return 0, fmt.Errorf("MummerMatchBp: %w", e)
	}
	bl, del, e := Mummer(
		[]fastats.FaEntry{wp.Fa1},
		[]fastats.FaEntry{wp.Fa2},
	)
	if e != nil {
		return h(e)
	}
	defer del()

	// bl.Stderr = os.Stderr

	var b strings.Builder
	bl.Stdout = &b
	e = bl.Run()
	if e != nil {
		return h(fmt.Errorf("bl.Run(): %w", e))
	}

	var stats []MummerMatch
	stats, e = AppendMummerMatches(strings.NewReader(b.String()), stats)
	if e != nil {
		return h(e)
	}
	var sum int64
	for _, stat := range stats {
		sum += stat.Length
	}
	return sum, nil
}
