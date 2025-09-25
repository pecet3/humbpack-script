package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/pecet3/hmbk-script/lexer"
	"github.com/pecet3/hmbk-script/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			printParserErr(out, p.Errors())
			continue
		}

		fmt.Println(program.String())
	}
}

func printParserErr(out io.Writer, errors []string) {
	for _, err := range errors {
		io.WriteString(out, err+"\n")
	}
}
