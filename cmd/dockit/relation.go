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
	Usage:   "transpose <file> [<sheet>]",
	Handler: &TransposeCommand{},
}

type TransposeCommand struct{}

func (c TransposeCommand) Run(args []string) error {
	set := cli.NewFlagSet("transpose")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 2 {
		return cli.ErrUsage
	}
	return withSheet(set.Arg(0), set.Arg(1), func(sh grid.View) error {
		var (
			view = grid.NewTransposedView(sh)
			rd   = cli.NewTableRenderer(cli.Stdout)
		)
		rd.Render(sheet2Table(view, false))
		return nil
	})

}

var joinCmd = cli.Command{
	Name:    "join",
	Summary: "Perform a join on two sheets",
	Usage:   "join <wb1> <sheet1> <key1> <wb2> <sheet2> <key2>",
	Handler: &JoinCommand{},
}

type JoinCommand struct{}

func (c JoinCommand) Run(args []string) error {
	set := cli.NewFlagSet("join")
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

	view := gridx.Join(sh1, sh2, sel1, sel2)

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
