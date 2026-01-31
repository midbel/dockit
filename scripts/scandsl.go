package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/formula/eval"
	"github.com/midbel/dockit/formula/op"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	ps := eval.NewParser(eval.ScriptGrammar())
	expr, err := ps.Parse(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	script, ok := expr.(eval.Script)
	if !ok {
		return
	}
	for _, e := range script.Body {
		str := eval.DumpExpr(e)
		fmt.Println(str)
	}
}

func scan(r io.Reader) {
	scan, err := eval.Scan(r, eval.ModeScript)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == op.EOF {
			break
		}
	}
}
