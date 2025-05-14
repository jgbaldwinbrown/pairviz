package pairviz

import (
	"os"
	"bufio"
	"log"
)

func FullPairvizMulti() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	flags := GetFlags()
	if e := WriteWindowsMulti(w, WinStatsMulti(flags, os.Stdin)); e != nil {
		log.Fatal(e)
	}
}
