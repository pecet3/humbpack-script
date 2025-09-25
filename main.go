package main

import (
	"os"

	"github.com/pecet3/hmbk-script/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
