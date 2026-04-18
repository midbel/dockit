package main

import (
	"github.com/midbel/cli"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/gridx"
	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/workbook"
)

var transposeCmd = cli.Command{
	Name:    "transpose",
	Summary: "Transpose rows and columns in a sheet",
	Help: `Arguments:
  file    path input to file
  sheet   name of sheet - if not provided active will be used

Options:
  -f <file>       path where transposed sheet will be written
  -c <columns>    selection of columns from input sheet to be transposed`,
	Usage:   "transpose [-f <output>] [-c <columns>] <file> [<sheet>]",
	Handler: &TransposeCommand{},
}

type TransposeCommand struct {
	OutFile string
	Columns layout.Selection
}

func (c TransposeCommand) Run(args []string) error {
	set := cli.NewFlagSet("transpose")
	set.StringVar(&c.OutFile, "f", "", "Write result to file")
	set.Func("c", "Selected columns", func(str string) error {
		sel, err := layout.SelectionFromString(str)
		if err == nil {
			c.Columns = sel
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() < 1 {
		return cli.ErrUsage
	}
	return withSheet(set.Arg(0), set.Arg(1), func(sh grid.View) error {
		view := c.createView(sh)
		return c.writeView(view)
	})
}

func (c TransposeCommand) writeView(view grid.View) error {
	if c.OutFile != "" {
		return workbook.WriteView(view, c.OutFile)
	}
	rd := cli.NewTableRenderer(cli.Stdout)
	rd.Render(sheet2Table(view, false))
	return nil
}

func (c TransposeCommand) createView(view grid.View) grid.View {
	if c.Columns != nil {
		view = grid.NewProjectView(view, c.Columns)
		view = grid.NewTransposedView(view)
	}
	return view
}

var groupCmd = cli.Command{
	Name:    "group",
	Summary: "",
	Help:    "",
	Usage:   "group <file> <sheet> <columns> <aggr> <col> [<aggr> <col>...]",
	Handler: &GroupCommand{},
}

type GroupCommand struct{}

func (c GroupCommand) Run(args []string) error {
	set := cli.NewFlagSet("group")
	if err := set.Parse(args); err != nil {
		return err
	}

	return withSheet(set.Arg(0), set.Arg(1), func(view grid.View) error {
		keys, err := layout.SelectionFromString(set.Arg(2))
		if err != nil {
			return err
		}
		var list []gridx.Aggr
		for i := 3; i < set.NArg(); i += 2 {
			a, err := gridx.CreateAggr(set.Arg(i+1), set.Arg(i))
			if err != nil {
				return err
			}
			list = append(list, *a)
		}
		v, err := gridx.Group(view, keys, list)
		if err != nil {
			return err
		}

		rd := cli.NewTableRenderer(cli.Stdout)
		rd.Render(sheet2Table(v, false))
		return nil
	})
}

var joinCmd = cli.Command{
	Name:    "join",
	Summary: "Perform a join on two sheets",
	Help: `Arguments:
  wb1       path to left file
  sheet1    name of sheet from left file
  key1      selection of columns used for the join
  wb2       path to right file
  sheet2    name of sheet from right file
  key2      selection of columns used for the join

Options:
  -f <file>       path where joined sheets will be written
  -c <columns>    selection of columns from joined sheets to be selected`,
	Usage:   "join [-f <output>] [-c <columns>] [-h|--help] <wb1> <sheet1> <key1> <wb2> <sheet2> <key2>",
	Handler: &JoinCommand{},
}

type JoinCommand struct {
	OutFile string
	Columns layout.Selection
}

func (c JoinCommand) Run(args []string) error {
	set := cli.NewFlagSet("join")
	set.StringVar(&c.OutFile, "f", "", "Write result to file")
	set.Func("c", "Selected columns", func(str string) error {
		sel, err := layout.SelectionFromString(str)
		if err == nil {
			c.Columns = sel
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 6 {
		return cli.ErrUsage
	}
	sh1, err := c.openSheet(set.Arg(0), set.Arg(1))
	if err != nil {
		return err
	}
	sh2, err := c.openSheet(set.Arg(3), set.Arg(4))
	if err != nil {
		return err
	}
	sel1, err := layout.SelectionFromString(set.Arg(2))
	if err != nil {
		return err
	}
	sel2, err := layout.SelectionFromString(set.Arg(5))
	if err != nil {
		return err
	}

	view := c.createView(gridx.Join(sh1, sh2, sel1, sel2))
	return c.writeView(view)
}

func (c JoinCommand) createView(view grid.View) grid.View {
	if c.Columns != nil {
		view = grid.NewProjectView(view, c.Columns)
	}
	return view
}

func (c JoinCommand) writeView(view grid.View) error {
	if c.OutFile != "" {
		return workbook.WriteView(view, c.OutFile)
	}
	rd := cli.NewTableRenderer(cli.Stdout)
	rd.Render(sheet2Table(view, false))
	return nil
}

func (c JoinCommand) openSheet(file, sheet string) (grid.View, error) {
	wb, err := workbook.Open(file)
	if err != nil {
		return nil, err
	}
	if sheet != "" {
		return wb.Sheet(sheet)
	}
	return wb.ActiveSheet()
}
