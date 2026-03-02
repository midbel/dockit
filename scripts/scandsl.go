package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/formula/op"
	"github.com/midbel/dockit/formula/parse"
)

func main() {
	var (
		lex  = flag.Bool("x", false, "lex")
		mode = flag.String("m", "", "mode")
	)
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	if *lex {
		err = scanReader(r, *mode)
	} else {
		err = parseReader(r, *mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func parseReader(r io.Reader, mode string) error {
	scanHint := parse.ScanScript
	if mode == "formula" {
		scanHint = parse.ScanFormula
	}
	scan, err := parse.Scan(r, scanHint)
	if err != nil {
		return err
	}
	ps, err := parse.NewParser(scan)
	if err != nil {
		return err
	}
	expr, err := ps.Parse()
	if err != nil {
		return err
	}
	script, ok := expr.(parse.Script)
	if !ok {
		return nil
	}
	for _, e := range script.Body {
		str := parse.DumpExpr(e)
		fmt.Println(str)
	}
	return nil
}

func scanReader(r io.Reader, mode string) error {
	scanHint := parse.ScanScript
	if mode == "formula" {
		scanHint = parse.ScanFormula
	}
	scan, err := parse.Scan(r, scanHint)
	if err != nil {
		return err
	}

	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == op.EOF || tok.Type == op.Invalid {
			break
		}
	}
	return nil
}
