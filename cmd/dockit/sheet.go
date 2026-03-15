package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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

var depsCmd = cli.Command{
	Name:    "deps",
	Summary: "Print dependencies information",
	Usage:   "deps <file>",
	Handler: &DepCommand{},
}

var auditCmd = cli.Command{
	Name:    "audit",
	Summary: "",
	Usage:   "audit <file> <sheet>",
	Handler: &AuditCommand{},
}

var lockCmd = cli.Command{
	Name:    "lock",
	Summary: "Lock a sheet or all in a spreadsheet",
	Usage:   "lock <file> [<sheet>]",
	Handler: &LockCommand{},
}

var unlockCmd = cli.Command{
	Name:    "unlock",
	Summary: "Unlock a sheet or all from a spreadsheet",
	Usage:   "unlock <file> [<sheet>]",
	Handler: &UnlockCommand{},
}

type DepCommand struct{}

func (c DepCommand) Run(args []string) error {
	set := cli.NewFlagSet("deps")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 2 {
		return cli.ErrUsage
	}
	return withSheet(set.Arg(0), set.Arg(1), func(sh grid.View) error {
		return nil
	})
}

type AuditCommand struct {
	Function bool
}

func (c AuditCommand) Run(args []string) error {
	set := cli.NewFlagSet("audit")
	set.BoolVar(&c.Function, "f", false, "function")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 2 {
		return cli.ErrUsage
	}
	return withSheet(set.Arg(0), set.Arg(1), func(sh grid.View) error {
		if c.Function {
			return c.auditFunction(sh)
		}
		return c.audit(sh)
	})
}

func (c AuditCommand) auditFunction(view grid.View) error {
	var (
		stats = grid.TopComplexFormulas(view, 10)
		tbl   cli.Table
	)
	tbl.Headers = []string{
		"location",
		"formula",
		"references",
		"complexity",
		"max depth",
	}
	for _, s := range stats {
		formula := s.Formula
		if len(formula) >= 60 {
			formula = formula[60:]
		}
		row := []string{
			s.Position.String(),
			formula,
			strconv.Itoa(len(s.Deps)),
			strconv.Itoa(s.Complexity),
			strconv.Itoa(s.MaxDepth),
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	rd := cli.NewTableRenderer(os.Stdout)
	rd.Render(tbl)
	return nil
}

func (c AuditCommand) audit(view grid.View) error {
	var (
		stats = grid.AnalyzeView(view)
		tbl1  cli.Table
		tbl2  cli.Table
		tbl3  cli.Table
		tbl4  cli.Table
	)
	tbl1.Title = fmt.Sprintf("sheet: %s", stats.Name)
	tbl1.Rows = [][]string{
		{"Cells", strconv.Itoa(stats.Cells)},
		{"Formulas", strconv.Itoa(stats.Formulas)},
		{"Errors", strconv.Itoa(stats.Errors)},
		{"Constants", strconv.Itoa(stats.Constants)},
	}
	if len(stats.Builtins) > 0 {
		tbl2.Title = "Top builtins"
		for _, t := range stats.TopBuiltins(5) {
			tbl2.Rows = append(tbl2.Rows, []string{
				strings.ToLower(t.Item),
				strconv.Itoa(t.Count),
			})
		}
	}
	if len(stats.Refs) > 0 {
		tbl3.Title = "Top Cells"
		for _, t := range stats.TopRefs(5) {
			tbl3.Rows = append(tbl3.Rows, []string{
				t.Item.String(),
				strconv.Itoa(t.Count),
			})
		}
	}

	tbl4.Title = "Most complexity formula"
	tbl4.Rows = [][]string{
		{"max depth", strconv.Itoa(stats.MaxDepth)},
		{"complexity", strconv.Itoa(stats.Complexity)},
	}

	rd := cli.NewTableRenderer(os.Stdout)
	rd.Render(tbl1)
	rd.Empty()
	if len(tbl2.Rows) > 0 {
		rd.Render(tbl2)
		rd.Empty()
	}
	if len(tbl3.Rows) > 0 {
		rd.Render(tbl3)
		rd.Empty()
	}
	rd.Render(tbl4)
	return nil
}

type LockCommand struct{}

func (c LockCommand) Run(args []string) error {
	set := cli.NewFlagSet("lock")
	if err := set.Parse(args); err != nil {
		return err
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		k, ok := wb.(interface{ LockSheet(string) })
		if !ok {
			return nil
		}
		for i := 1; i < set.NArg(); i++ {
			k.LockSheet(set.Arg(i))
		}
		return nil
	})
}

type UnlockCommand struct{}

func (c UnlockCommand) Run(args []string) error {
	set := cli.NewFlagSet("unlock")
	if err := set.Parse(args); err != nil {
		return err
	}
	return updateFile(set.Arg(0), func(wb grid.File) error {
		k, ok := wb.(interface{ UnLockSheet(string) })
		if !ok {
			return nil
		}
		for i := 1; i < set.NArg(); i++ {
			k.UnLockSheet(set.Arg(i))
		}
		return nil
	})
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
	if err := wb.Reload(); err != nil && !errors.Is(err, grid.ErrSupported) {
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
