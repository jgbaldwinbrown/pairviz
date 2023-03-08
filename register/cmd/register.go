package main

import (
	"os"
	"bufio"
	"github.com/jgbaldwinbrown/pairviz/register/pkg"
)

func main() {
	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	e := register.Run(os.Stdin, stdout)
	if e != nil { panic(e) }
}
