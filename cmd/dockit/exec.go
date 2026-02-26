package main

import (
	"fmt"
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/formula/repr"
)

type RunCommand struct{}

func (c RunCommand) Run(args []string) error {
	return nil
}

type DumpCommand struct{}

func (c DumpCommand) Run(args []string) error {
	set := cli.NewFlagSet("dump")
	if err := set.Parse(args); err != nil {
		return err
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		return err
	}
	defer r.Close()

	root, err := repr.Inspect(r)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", root)
	return nil
}
