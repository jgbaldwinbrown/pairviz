package pairviz

import (
	"fmt"
	"bufio"
	"log"
	"os"
)

func FullPairvizMultiRegion() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	flags := GetFlags()
	if flags.Region == "" {
		log.Fatal(fmt.Errorf("Missing -r"))
	}
	regions, regionsMap, e := BedfileToRegions(flags.Region)
	if e != nil {
		log.Fatal(e)
	}
	if e := WriteRegionsMulti(w, RegionStatsMulti(flags, regions, regionsMap, os.Stdin)); e != nil {
		log.Fatal(e)
	}
}
