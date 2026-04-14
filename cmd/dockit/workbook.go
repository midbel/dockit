package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/grid/builtins"
	"github.com/midbel/dockit/workbook"
	"github.com/midbel/textwrap"
)

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "Display metadata, sheet names of a spreadsheet file",
	Help: `Arguments:
  file    path to input file.

Options:
  -f <format>    force to use the given format
  -p <pattern>   use pattern to extract columns from log file`,
	Usage:   "info [-f <format>] [-p <pattern>] [-h|--help] <file>",
	Handler: &GetInfoCommand{},
}

var mergeCmd = cli.Command{
	Name:    "merge",
	Summary: "Consolidate multiple spreadsheet files into a single workbooks",
	Help: `Arguments:
  file    path to input file(s).

Options:
  -f <file>    write merge result to given file
  -r           remove input file(s)
  -c           resync values before merge`,
	Usage:   "merge [-f <file>] [-r] [-c] <file...>",
	Handler: &MergeCommand{},
}

var formatCmd = cli.Command{
	Name:    "format",
	Summary: "List supported spreadsheet formats",
	Usage:   "format [-h|--help]",
	Handler: &FormatCommand{},
}

var builtinsCmd = cli.Command{
	Name:    "builtins",
	Summary: "Display list of supported builtins",
	Help: `Arguments:
  name    builtin's name to get detailled info

Options:
  -c <category>    list builtins from given category only`,
	Usage:   "builtins [-c <category>] [-h|--help] [<name>]",
	Handler: &GetBuiltinCommand{},
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
		if err := wb.Sync(); err != nil && !errors.Is(err, grid.ErrSupported) {
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
	Format  string
	Pattern string
}

func (c GetInfoCommand) Run(args []string) error {
	set := cli.NewFlagSet("info")
	set.StringVar(&c.Format, "f", "", "format")
	set.StringVar(&c.Pattern, "p", "", "pattern")
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
	tbl.Headers = []string{"sheet", "active", "locked", "visible", "rows", "columns"}
	for _, i := range file.Infos() {
		r := []string{
			i.Name,
			cli.MarkBool(i.Active),
			cli.MarkBool(i.Protected),
			cli.MarkBool(!i.Hidden),
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

type GetBuiltinCommand struct {
	Category string
}

func (c GetBuiltinCommand) Run(args []string) error {
	set := cli.NewFlagSet("builtins")
	set.StringVar(&c.Category, "c", "", "List functions of given category")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() >= 1 {
		return c.printHelp(set.Arg(0))
	} else {
		return c.printList()
	}
}

func (c GetBuiltinCommand) printHelp(name string) error {
	return builtins.Help(cli.Stdout, name)
}

func (c GetBuiltinCommand) printList() error {
	var (
		list = builtins.List()
		tbl  cli.Table
	)
	slices.SortFunc(list, func(b1, b2 builtins.Builtin) int {
		z := strings.Compare(b1.Category, b2.Category)
		if z == 0 {
			return strings.Compare(b1.Name, b2.Name)
		}
		return z
	})
	tbl.Headers = []string{"Name", "Description", "Category", "Parameter", "Openxml", "Opendoc"}
	for _, b := range list {
		if c.Category != "" && b.Category != c.Category {
			continue
		}
		r := []string{
			b.Name,
			textwrap.Shorten(b.Desc, 48),
			b.Category,
			strconv.Itoa(len(b.Params)),
			cli.MarkBool(b.OxmlSupported()),
			cli.MarkBool(b.OdsSupported()),
		}
		tbl.Rows = append(tbl.Rows, r)
	}
	rd := cli.NewTableRenderer(cli.Stdout)
	rd.WithLineNumbers = true
	rd.Render(tbl)
	return nil
}
