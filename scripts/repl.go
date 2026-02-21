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
	dir   := flag.String("d", "", "context directory")
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
	_, err = engine.Exec(r, env.Empty())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
