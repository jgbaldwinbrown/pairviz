package windif

import (
	"fmt"
	"github.com/jgbaldwinbrown/iter"
	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/fastats/pkg"
)

func Winpairs(paths []string) *iter.Iterator[Winpair] {
	return &iter.Iterator[Winpair]{Iteratef: func(yield func(Winpair) error) error {
		if len(paths) != 2 {
			return fmt.Errorf("Winpairs: len(paths) %d != 2", len(paths))
		}
		r0, e := csvh.OpenMaybeGz(paths[0])
		if e != nil {
			return e
		}
		defer r0.Close()
		r1, e := csvh.OpenMaybeGz(paths[1])
		if e != nil {
			return e
		}
		defer r1.Close()

		f0 := fastats.ParseFasta(r0)
		f1 := fastats.ParseFasta(r1)

		p1 := iter.Pull[fastats.FaEntry](f1, 0)
		defer p1.Close()

		return f0.Iterate(func(fa0 fastats.FaEntry) error {
			fa1, e := p1.Next()
			if e != nil {
				return e
			}
			return yield(Winpair{fa0, fa1})
		})
	}}
}
