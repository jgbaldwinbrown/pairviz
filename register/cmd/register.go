package main

import (
	"os"
	"bufio"
	"github.com/jgbaldwinbrown/pairviz/register/pkg"
	"flag"
)

type Flags struct {
	Maxdist int
}

func main() {
	var f Flags
	flag.IntVar(&f.Maxdist, "m", 30000, "Maximum distance to plot")
	flag.Parse()

	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	e := register.Run(int64(f.Maxdist), os.Stdin, stdout)
	if e != nil { panic(e) }
}
