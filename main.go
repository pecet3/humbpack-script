package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pecet3/hmbk-script/evaluation"
	"github.com/pecet3/hmbk-script/lexer"
	"github.com/pecet3/hmbk-script/object"
	"github.com/pecet3/hmbk-script/parser"
	"github.com/pecet3/hmbk-script/repl"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		fileName := args[0]
		fileNameLower := strings.ToLower(fileName)
		if !strings.Contains(fileNameLower, ".hmbk") {
			fmt.Println("Wrong file name, it must have a .hmbk extension")
			return
		}
		data, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Błąd wczytania pliku: %s\n", err)
			os.Exit(1)
		}

		env := object.NewEnvironment()
		l := lexer.New(string(data))
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			for _, err := range p.Errors() {
				fmt.Fprintf(os.Stderr, "Parser error: %s\n", err)
			}
			os.Exit(1)
		}

		evaluated := evaluation.Eval(program, env)

		if evaluated != nil {
			_, isErr := evaluated.(*object.GlobalError)
			if isErr {
				fmt.Fprintf(os.Stderr, "%s\n", evaluated.Inspect())
				os.Exit(1)
			}

		}
	} else {
		repl.Start(os.Stdin, os.Stdout)
	}
}
