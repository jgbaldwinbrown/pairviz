package windif

import (
	"os"
	"io"
	"github.com/jgbaldwinbrown/fastats/pkg"
	"github.com/jgbaldwinbrown/csvh"
	"os/exec"
	"path/filepath"
)

func Blast(sub, quer []fastats.FaEntry) (cmd *exec.Cmd, del func(), err error) {
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

	cmd = exec.Command("blastn", "-subject", spath, "-query", qpath, "-outfmt", "6")
	return cmd, func(){os.RemoveAll(dir)}, nil
}

type BlastStat struct {
	Qseqid string
	Sseqid string
	Pident float64
	Length int64
	Mismatch int64
	Gapopen int64
	Qstart int64
	Qend int64
	Sstart int64
	Send int64
	Evalue float64
	Bitscore int64
}

func AppendBlastStats(r io.Reader, stats []BlastStat) ([]BlastStat, error) {
	cr := csvh.CsvIn(r)
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return nil, e
		}
		var s BlastStat
		_, e := csvh.Scan(l,
			&s.Qseqid, &s.Sseqid,
			&s.Pident, &s.Length,
			&s.Mismatch, &s.Gapopen,
			&s.Qstart, &s.Qend,
			&s.Sstart, &s.Send,
			&s.Evalue, &s.Bitscore,
		)
		if e != nil {
			return nil, e
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func BestBitscore(wp Winpair) (int64, error) {
	ec := make(chan error)
	bl, del, e := Blast(
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

	var stats []BlastStat
	stats, e = AppendBlastStats(pr, stats)
	if e != nil {
		return 0, e
	}
	var best int64
	for _, stat := range stats {
		if best < stat.Bitscore {
			best = stat.Bitscore
		}
	}
	return best, nil
}
