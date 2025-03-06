package dsamp

import (
	"fmt"
	"math/rand"
	"flag"
	"os"
	"bufio"
	"log"
)

type DownsampleRandomFlags struct {
	Perc float64
	Seed int64
}

func RunDownsampleRandom() {
	var f DownsampleRandomFlags
	flag.Float64Var(&f.Perc, "p", 1.0, "Percentage of data to keep")
	flag.Int64Var(&f.Seed, "s", 0, "Random seed")
	flag.Parse()

	rng := rand.New(rand.NewSource(f.Seed))

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		err := w.Flush()
		if (err != nil) {
			log.Fatal(err)
		}
	}()

	s := bufio.NewScanner(os.Stdin)
	s.Buffer([]byte{}, 1e18)
	for s.Scan() {
		if s.Err() != nil {
			log.Fatal(s.Err())
		}
		if len(s.Text()) > 0 && s.Text()[0] == '#' {
			continue
		}
		r := rng.Float64()
		if r >= f.Perc {
			continue
		}
		if _, e := fmt.Fprintln(w, s.Text()); e != nil {
			log.Fatal(e)
		}
	}
}
