package main

import (
	"fmt"
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/workbook"
)

var mergeCmd = cli.Command{
	Name:    "merge",
	Summary: "merge multiple files in one spreadsheet",
	Usage:   "merge [-o] <file1> <file2> [...<fileN>]",
	Handler: &MergeCommand{},
}

var formatCmd = cli.Command{
	Name:    "format",
	Summary: "retrieve list of supported formats",
	Usage:   "",
	Handler: &FormatCommand{},
}

type MergeCommand struct{}

func (c MergeCommand) Run(args []string) error {
	var (
		set    = cli.NewFlagSet("merge")
		file   = set.String("f", "merge.xlsx", "write merge result into given file")
		remove = set.Bool("r", false, "remove files merged")
	)
	if err := set.Parse(args); err != nil {
		return err
	}
	if err := c.mergeFiles(*file, set.Args()); err != nil {
		return err
	}
	if *remove {
		c.removeFiles(set.Args())
	}
	return nil
}

func (c MergeCommand) mergeFiles(file string, sources []string) error {
	return workbook.Merge(file, sources)
}

func (c MergeCommand) removeFiles(files []string) {
	for _, f := range files {
		os.Remove(f)
	}
}

type FormatCommand struct{}

func (c FormatCommand) Run(args []string) error {
	set := cli.NewFlagSet("merge")
	if err := set.Parse(args); err != nil {
		return err
	}
	for _, n := range workbook.Formats() {
		fmt.Fprintln(os.Stdout, n)
	}
	return nil
}
