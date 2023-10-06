package pairviz

import (
	"os"
	"bufio"
)

func FullPairviz() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	flags := GetFlags()
	if flags.Chromosome {
		FprintChromStats(w, ChromosomeStats(flags, os.Stdin))
	} else if flags.Region != "" {
		regions, err := GetRegionStats(flags, os.Stdin)
		if err != nil {panic(err)}
		FprintRegionStats(w, regions)
	} else {
		FprintWinStats(w, WinStats(flags, os.Stdin), flags.SeparateGenomes, flags.ReadLen, flags.JsonOut)
	}
}
