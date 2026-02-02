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
	lex := flag.Bool("x", false, "lex")
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	if *lex {
		err = scan(r)
	} else {
		err = parse(r)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func parse(r io.Reader) error {
	ps := eval.NewParser(eval.ScriptGrammar())
	expr, err := ps.Parse(r)
	if err != nil {
		return err
	}
	script, ok := expr.(eval.Script)
	if !ok {
		return nil
	}
	for _, e := range script.Body {
		str := eval.DumpExpr(e)
		fmt.Println(str)
	}
	return nil
}

func scan(r io.Reader) error {
	scan, err := eval.Scan(r, eval.ModeScript)
	if err != nil {
		return err
	}

	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == op.EOF {
			break
		}
	}
	return nil
}
