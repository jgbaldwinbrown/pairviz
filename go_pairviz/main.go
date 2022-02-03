package main

import (
	"os"
)

func main() {
	flags := GetFlags()
	if flags.Chromosome {
		PrintChromStats(ChromosomeStats(flags, os.Stdin))
	} else if flags.Region != "" {
		regions, err := GetRegionStats(flags, os.Stdin)
		if err != nil {panic(err)}
		PrintRegionStats(regions)
	} else {
		PrintWinStats(WinStats(flags, os.Stdin))
	}
}
