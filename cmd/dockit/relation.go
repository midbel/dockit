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
	Usage:   "join sheet sheet",
	Handler: &JoinCommand{},
}

type JoinCommand struct{}

func (c JoinCommand) Run(args []string) error {
	set := cli.NewFlagSet("join")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 2 {
		return cli.ErrUsage
	}
	sh1, err := c.openSheet(set.Arg(0))
	if err != nil {
		return err
	}
	sh2, err := c.openSheet(set.Arg(1))
	if err != nil {
		return err
	}
	var (
		cols = layout.SelectSingle(1)
		view = gridx.Join(sh1, sh2, cols, cols)
	)

	rd := cli.NewTableRenderer(cli.Stdout)
	rd.Render(sheet2Table(view, false))
	return nil
}

func (c JoinCommand) openSheet(file string) (grid.View, error) {
	wb, err := workbook.Open(file)
	if err != nil {
		return nil, err
	}
	return wb.ActiveSheet()
}
