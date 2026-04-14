package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/grid"
)

var depsCmd = cli.Command{
	Name:    "deps",
	Summary: "Print dependencies information",
	Usage:   "deps <file>",
	Handler: &DepCommand{},
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

var auditCmd = cli.Command{
	Name:    "audit",
	Summary: "",
	Usage:   "audit <file> <sheet>",
	Handler: &AuditCommand{},
}


type AuditCommand struct {
	Function bool
	Limit    int
}

func (c AuditCommand) Run(args []string) error {
	set := cli.NewFlagSet("audit")
	set.BoolVar(&c.Function, "f", false, "function")
	set.IntVar(&c.Limit, "n", 10, "number of results")
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
		stats = grid.TopComplexFormulas(view, c.Limit)
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
		for _, t := range stats.TopBuiltins(c.Limit) {
			tbl2.Rows = append(tbl2.Rows, []string{
				strings.ToLower(t.Item),
				strconv.Itoa(t.Count),
			})
		}
	}
	if len(stats.Refs) > 0 {
		tbl3.Title = "Top Cells"
		for _, t := range stats.TopRefs(c.Limit) {
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
