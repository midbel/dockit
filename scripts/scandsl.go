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

	// scan(r)
	_, err = formula.Exec(r, formula.Empty())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
