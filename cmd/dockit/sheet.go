package main

import (
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/workbook"
)

var addCmd = cli.Command{
	Name:    "add",
	Alias:   slx.Make("append"),
	Summary: "Import specific sheets from a source spreadsheet into a target file",
	Usage:   "add <target> <source> <sheet> [<sheet>...]",
	Handler: &AddCommand{},
}

var dropCmd = cli.Command{
	Name:    "drop",
	Alias:   slx.Make("remove", "rm"),
	Summary: "Delete one or more sheets from a spreadsheet",
	Usage:   "drop <file> <sheet> [<sheet>...]",
	Handler: &DropCommand{},
}

var renameCmd = cli.Command{
	Name:    "rename",
	Summary: "Change the name of a specific sheet within a file",
	Usage:   "rename <file> <source> <target>",
	Handler: &RenameCommand{},
}

var copyCmd = cli.Command{
	Name:    "copy",
	Alias:   slx.Make("cp"),
	Summary: "Duplicate a sheet within its original file",
	Usage:   "copy <file> <sheet>",
	Handler: &CopyCommand{},
}

var printCmd = cli.Command{
	Name:    "print",
	Summary: "Print content of a sheet on stdout",
	Usage:   "print <file> [<sheet>]",
	Handler: &PrintCommand{},
}

type AddCommand struct{}

func (c AddCommand) Run(args []string) error {
	set := cli.NewFlagSet("add")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() <= 2 {
		return cli.ErrUsage
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		other, err := workbook.Open(set.Arg(1))
		if err != nil {
			return err
		}
		for i := 2; i < set.NArg(); i++ {
			sh, err := other.Sheet(set.Arg(i))
			if err != nil {
				return err
			}
			if err := wb.AppendSheet(sh); err != nil {
				return err
			}
		}
		return nil
	})
}

type DropCommand struct{}

func (c DropCommand) Run(args []string) error {
	set := cli.NewFlagSet("drop")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() <= 1 {
		return cli.ErrUsage
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		for i := 1; i < set.NArg(); i++ {
			if err := wb.RemoveSheet(set.Arg(i)); err != nil {
				return err
			}
		}
		return nil
	})
}

type CopyCommand struct{}

func (c CopyCommand) Run(args []string) error {
	set := cli.NewFlagSet("copy")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 3 {
		return cli.ErrUsage
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		return wb.Copy(set.Arg(1), set.Arg(2))
	})
}

type RenameCommand struct{}

func (c RenameCommand) Run(args []string) error {
	set := cli.NewFlagSet("rename")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 3 {
		return cli.ErrUsage
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		return wb.Rename(set.Arg(1), set.Arg(2))
	})
}

type PrintCommand struct {
	Format    string
	Pattern   string
	Delimiter string
}

func (c PrintCommand) Run(args []string) error {
	set := cli.NewFlagSet("print")
	set.StringVar(&c.Format, "f", "", "format")
	set.StringVar(&c.Pattern, "p", "", "pattern")
	set.StringVar(&c.Delimiter, "d", "", "format")
	if err := set.Parse(args); err != nil {
		return err
	}
	wb, err := c.openFile(set.Arg(0))
	if err != nil {
		return err
	}
	if err := wb.Reload(); err != nil {
		return err
	}
	var sheet grid.View
	if set.NArg() == 1 {
		sheet, err = wb.ActiveSheet()
	} else {
		sheet, err = wb.Sheet(set.Arg(1))
	}
	if err != nil {
		return err
	}
	rd := cli.NewTableRenderer(os.Stdout)
	rd.Render(sheet2Table(sheet))
	return nil
}

func (c PrintCommand) openFile(file string) (grid.File, error) {
	if c.Format == "log" {
		return flat.OpenLog(file, c.Pattern)
	}
	return workbook.OpenFormat(file, c.Format)
}

func sheet2Table(sheet grid.View) cli.Table {
	var (
		t cli.Table
		i int
	)
	for r := range sheet.Rows() {
		row := make([]string, 0, len(r))
		for _, v := range r {
			row = append(row, v.String())
		}
		if i == 0 {
			t.Headers = row
		} else {
			t.Rows = append(t.Rows, row)
		}
		i++
	}
	return t
}
