package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/dockit/oxml"
)

func main() {
	flag.Parse()

	f, err := oxml.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, s := range f.SheetNames() {
		fmt.Println(s)
	}
}
