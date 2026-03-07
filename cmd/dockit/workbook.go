package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/workbook"
)

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "Display metadata, sheet names of a spreadsheet file",
	Usage:   "info [-a] <spreadsheet>",
	Handler: &GetInfoCommand{},
}

var mergeCmd = cli.Command{
	Name:    "merge",
	Summary: "Consolidate multiple spreadsheet files into a single workbooks",
	Usage:   "merge [-o] [-r] <file1> <file2> [...<fileN>]",
	Handler: &MergeCommand{},
}

var formatCmd = cli.Command{
	Name:    "format",
	Summary: "Print list of supported spreadsheet like formats",
	Usage:   "List all compatible spreadsheet file format",
	Handler: &FormatCommand{},
}

type MergeCommand struct{}

func (c MergeCommand) Run(args []string) error {
	var (
		set    = cli.NewFlagSet("merge")
		file   = set.String("f", "merge.xlsx", "write merge result into given file")
		remove = set.Bool("r", false, "remove files merged")
		reload = set.Bool("c", false, "recompute all values in final file")
	)
	if err := set.Parse(args); err != nil {
		return err
	}
	wb, err := c.mergeFiles(*file, set.Args())
	if err != nil {
		return err
	}

	if *reload {
		if err := wb.Reload(); err != nil {
			return err
		}
	}
	if err := c.writeFile(wb, *file); err != nil {
		return err
	}
	if *remove {
		c.removeFiles(set.Args())
	}
	return nil
}

func (c MergeCommand) mergeFiles(file string, sources []string) (grid.File, error) {
	return workbook.Merge(filepath.Ext(file), sources)
}

func (c MergeCommand) writeFile(wb grid.File, file string) error {
	return workbook.WriteFile(wb, file)
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

const infoPattern = "%d %s%s(%s): %d lines, %d columns - %s"

type GetInfoCommand struct {
	Format    string
	Pattern   string
	Delimiter string
}

func (c GetInfoCommand) Run(args []string) error {
	set := cli.NewFlagSet("info")
	set.StringVar(&c.Format, "f", "", "format")
	set.StringVar(&c.Pattern, "p", "", "pattern")
	set.StringVar(&c.Delimiter, "d", "", "format")
	if err := set.Parse(args); err != nil {
		return err
	}
	file, err := c.openFile(set.Arg(0))
	if err != nil {
		return err
	}
	var (
		tbl cli.Table
		rd  = cli.NewTableRenderer(os.Stdout)
	)
	tbl.Headers = []string{"sheet", "active", "locked", "visibility", "rows", "columns"}
	for _, i := range file.Infos() {
		r := []string{
			i.Name,
			strconv.FormatBool(i.Active),
			strconv.FormatBool(i.Protected),
			strconv.FormatBool(!i.Hidden),
			strconv.FormatInt(i.Size.Lines, 10),
			strconv.FormatInt(i.Size.Columns, 10),
		}
		tbl.Rows = append(tbl.Rows, r)
	}
	return rd.Render(tbl)
}

func (c GetInfoCommand) openFile(file string) (grid.File, error) {
	if c.Format == "log" {
		return flat.OpenLog(file, c.Pattern)
	}
	return workbook.OpenFormat(file, c.Format)
}
