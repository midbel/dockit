package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/midbel/dockit/doc"
	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/eval"
)

func main() {
	var (
		dir   = flag.String("d", "", "context directory")
		debug = flag.Bool("g", false, "debug mode")
	)
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *dir == "" {
		*dir = filepath.Dir(flag.Arg(0))
	}

	engine := eval.NewEngine(doc.NewLoader(*dir))
	if *debug {
		engine.SetPrintMode(eval.PrintDebug)
	}
	_, err = engine.Exec(r, env.Empty())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
