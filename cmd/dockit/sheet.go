package main

import (
	"errors"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/workbook"
)

var addCmd = cli.Command{
	Name:    "add",
	Alias:   slx.Make("append"),
	Summary: "Import sheets from a spreadsheet into a target file",
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
	Usage:   "print [-c <columns>] <file> [<sheet>]",
	Handler: &PrintCommand{},
}

var lockCmd = cli.Command{
	Name:    "lock",
	Summary: "Lock one or more sheets from a spreadsheet",
	Usage:   "lock <file> [<sheet>]",
	Handler: &LockCommand{},
}

var unlockCmd = cli.Command{
	Name:    "unlock",
	Summary: "Unlock one or more sheets from a spreadsheet",
	Usage:   "unlock <file> [<sheet>]",
	Handler: &UnlockCommand{},
}

type LockCommand struct{}

func (c LockCommand) Run(args []string) error {
	set := cli.NewFlagSet("lock")
	if err := set.Parse(args); err != nil {
		return err
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		var err error
		if set.NArg() <= 1 {
			c.lockFile(wb)
		} else {
			args := set.Args()
			err = c.lockSheets(wb, args[1:])
		}
		return err
	})
}

func (c LockCommand) lockFile(wb grid.File) {
	k, ok := wb.(interface{ Lock() })
	if !ok {
		return
	}
	k.Lock()
}

func (c LockCommand) lockSheets(wb grid.File, sheets []string) error {
	k, ok := wb.(interface{ LockSheet(string) error })
	if !ok {
		return nil
	}
	for _, sh := range sheets {
		err := k.LockSheet(sh)
		if err != nil {
			return err
		}
	}
	return nil
}

type UnlockCommand struct{}

func (c UnlockCommand) Run(args []string) error {
	set := cli.NewFlagSet("unlock")
	if err := set.Parse(args); err != nil {
		return err
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		var err error
		if set.NArg() <= 1 {
			c.unlockFile(wb)
		} else {
			args := set.Args()
			err = c.unlockSheets(wb, args[1:])
		}
		return err
	})
}

func (c UnlockCommand) unlockFile(wb grid.File) {
	k, ok := wb.(interface{ Unlock() })
	if !ok {
		return
	}
	k.Unlock()
}

func (c UnlockCommand) unlockSheets(wb grid.File, sheets []string) error {
	k, ok := wb.(interface{ UnlockSheet(string) error })
	if !ok {
		return nil
	}
	for _, sh := range sheets {
		err := k.UnlockSheet(sh)
		if err != nil {
			return err
		}
	}
	return nil
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
	Format  string
	Pattern string
	Columns layout.Selection
	Quoted  bool
	SkipErr bool
}

func (c PrintCommand) Run(args []string) error {
	set := cli.NewFlagSet("print")
	set.StringVar(&c.Format, "f", "", "format")
	set.StringVar(&c.Pattern, "p", "", "pattern")
	set.BoolVar(&c.Quoted, "q", false, "quoted")
	set.BoolVar(&c.SkipErr, "ignore-errors", false, "skip rows having error values")
	set.Func("c", "selected columns", func(str string) error {
		sel, err := layout.SelectionFromString(str)
		if err == nil {
			c.Columns = sel
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	sheet, err := c.openSheet(set.Arg(0), set.Arg(1))
	if err != nil {
		return err
	}
	if c.Columns != nil {
		sheet = grid.NewProjectView(sheet, c.Columns)
	}

	var rd cli.Renderer
	if c.Quoted {
		r := NewCsvRenderer(cli.Stdout)
		r.Quoted = c.Quoted
		rd = r
	} else {
		rd = cli.NewTableRenderer(cli.Stdout)

	}
	rd.Render(sheet2Table(sheet, c.SkipErr))
	return nil
}

func (c PrintCommand) openSheet(file, name string) (grid.View, error) {
	wb, err := c.openFile(file)
	if err != nil {
		return nil, err
	}
	if err := wb.Sync(); err != nil && !errors.Is(err, grid.ErrSupported) {
		return nil, err
	}
	var sheet grid.View
	if name == "" {
		sheet, err = wb.ActiveSheet()
	} else {
		sheet, err = wb.Sheet(name)
	}
	if err != nil {
		return nil, err
	}
	if c.Columns != nil {
		sheet = grid.NewProjectView(sheet, c.Columns)
	}
	return sheet, nil
}

func (c PrintCommand) openFile(file string) (grid.File, error) {
	if c.Format == "log" {
		return flat.OpenLog(file, c.Pattern)
	}
	return workbook.OpenFormat(file, c.Format)
}
