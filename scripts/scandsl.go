package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/dockit/formula/parse"
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
		err = scanReader(r)
	} else {
		err = parseReader(r)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func parseReader(r io.Reader) error {
	scan, err := parse.Scan(r, parse.ScanScript)
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

func scanReader(r io.Reader) error {
	scan, err := parse.Scan(r, parse.ScanScript)
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
