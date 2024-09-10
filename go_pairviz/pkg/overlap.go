package pairviz

import (
	"errors"
	"fmt"
)

var ErrImpossibleFacing = errors.New("Impossible facing")

// Check if the two reads in a pair overlap with each other
func PairOverlaps(p Pair, readlen int64) bool {
	ovl := PairOverlapsCore(p, readlen)
	// log.Printf("PairOverlaps: p: %v; readlen: %v; ovl: %v\n", p, readlen, ovl);
	return ovl
}

func PairOverlapsCore(p Pair, readlen int64) bool {
	if p.Read1.Parent != p.Read2.Parent {
		return false
	}
	if p.Read1.Chrom != p.Read2.Chrom {
		return false
	}

	f := p.Face()
	switch f {
	case Unknown:
		return false
	case Out:
		return p.Read1.Pos == p.Read2.Pos
	case Match:
		dist := p.AbsPosDist()
		return dist < readlen
	case In:
		dist := p.AbsPosDist()
		return dist < (readlen * 2)
	default:
		panic(fmt.Errorf("PairOverlaps: facing %v: %w", f, ErrImpossibleFacing))
	}

	return false
}
