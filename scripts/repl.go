package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/eval"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	engine := eval.NewEngine()
	_, err = engine.Exec(r, env.Empty())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
