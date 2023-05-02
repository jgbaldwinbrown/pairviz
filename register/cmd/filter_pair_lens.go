package main

import (
	"os"
	"bufio"
	"github.com/jgbaldwinbrown/pairviz/register/pkg"
)

func main() {
	args, e := register.GetFilterArgs()
	if e != nil { panic(e) }

	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	e = register.RunFilter(os.Stdin, stdout, args)
	if e != nil { panic(e) }
}
