package main

import (
	"github.com/midbel/cli"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/workbook"
	"github.com/midbel/dockit/internal/slx"
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
	Summary: "Change the display name of a specific sheet within a file",
	Usage:   "rename <file> <source> <target>",
	Handler: &RenameCommand{},
}

var copyCmd = cli.Command{
	Name:    "copy",
	Alias:   slx.Make("cp"),
	Summary: "Duplicate a sheet within its original file or transfer it to a new one",
	Usage:   "copy <file> <sheet>",
	Handler: &CopyCommand{},
}

type AddCommand struct{}

func (c AddCommand) Run(args []string) error {
	set := cli.NewFlagSet("add")
	if err := set.Parse(args); err != nil {
		return err
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
	return updateFile(set.Arg(0), func(wb grid.File) error {
		return wb.Rename(set.Arg(1), set.Arg(2))
	})
}
