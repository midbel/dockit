package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/formula"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	ps := formula.NewParser(formula.ScriptGrammar())
	expr, err := ps.Parse(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	script, ok := expr.(formula.Script)
	if !ok {
		return
	}
	for _, e := range script.Body {
		fmt.Println(e)
	}
}

func scan(r io.Reader) {
	scan, err := formula.Scan(r, formula.ModeScript)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == formula.EOF {
			break
		}
	}
}
