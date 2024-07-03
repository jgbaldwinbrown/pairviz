package windif

import (
	"fmt"
	"iter"
	"github.com/jgbaldwinbrown/csvh"
	"github.com/jgbaldwinbrown/fastats/pkg"
)

func Winpairs(paths []string) iter.Seq2[Winpair, error] {
	return func(yield func(Winpair, error) bool) {
		if len(paths) != 2 {
			yield(Winpair{}, fmt.Errorf("Winpairs: len(paths) %d != 2", len(paths)))
			return
		}
		r0, e := csvh.OpenMaybeGz(paths[0])
		if e != nil {
			yield(Winpair{}, e)
			return
		}
		defer r0.Close()
		r1, e := csvh.OpenMaybeGz(paths[1])
		if e != nil {
			yield(Winpair{}, e)
			return
		}
		defer r1.Close()

		f0 := fastats.ParseFasta(r0)
		f1 := fastats.ParseFasta(r1)

		p1 := iter.Pull[fastats.FaEntry](f1, 0)
		defer p1.Close()

		for fa0, e := range f0 {
			if e != nil {
				yield(Winpair{}, e)
				return
			}
			fa1, e := p1.Next()
			if e != nil {
				yield(Winpair{}, e)
				return
			}
			if ok := yield(Winpair{fa0, fa1}); !ok {
				return
			}
		}
	}
}
