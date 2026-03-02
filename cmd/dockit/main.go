package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/flat"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/oxml"
	"github.com/midbel/dockit/workbook"
)

var errFail = errors.New("fail")

var (
	summary = "dockit"
	help    = ""
)

func init() {
	workbook.Register(oxml.NewLoader())
	workbook.Register(flat.NewCommaLoader())
	workbook.Register(flat.NewTabLoader())
	workbook.Register(flat.NewSemicolonLoader())
	workbook.Register(flat.NewColonLoader())
}

func main() {
	var (
		set  = cli.NewFlagSet("dockit")
		root = prepare()
	)
	root.SetSummary(summary)
	root.SetHelp(help)
	if err := set.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			root.Help()
			os.Exit(2)
		}
	}
	err := root.Execute(set.Args())
	if err != nil {
		if s, ok := err.(cli.SuggestionError); ok && len(s.Others) > 0 {
			fmt.Fprintln(os.Stderr, "similar command(s)")
			for _, n := range s.Others {
				fmt.Fprintln(os.Stderr, "-", n)
			}
		}
		if !errors.Is(err, errFail) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func prepare() *cli.CommandTrie {
	root := cli.New()
	root.Register(slx.One("info"), &infoCmd)
	root.Register(slx.One("merge"), &mergeCmd)
	root.Register(slx.One("format"), &formatCmd)
	root.Register(slx.One("run"), &runCmd)
	root.Register(slx.One("dump"), &dumpCmd)
	root.Register(slx.One("query"), &queryCmd)
	root.Register(slx.One("extract"), &extractCmd)
	root.Register(slx.Make("sheet", "add"), &addCmd)
	root.Register(slx.Make("sheet", "drop"), &dropCmd)
	root.Register(slx.Make("sheet", "rename"), &renameCmd)
	root.Register(slx.Make("sheet", "copy"), &copyCmd)
	root.Register(slx.Make("sheet", "move"), &moveCmd)

	return root
}

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "get informations about sheets in given file",
	Usage:   "info [-a] <spreadsheet>",
	Handler: &GetInfoCommand{},
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
	set.StringVar(&c.Pattern, "p", "", "format")
	set.StringVar(&c.Delimiter, "d", "", "format")
	if err := set.Parse(args); err != nil {
		return err
	}
	file, err := c.openFile(set.Arg(0))
	if err != nil {
		return err
	}
	for j, i := range file.Infos() {
		c.printInfo(i, j)
	}
	return nil
}

func (c GetInfoCommand) openFile(file string) (grid.File, error) {
	if c.Format == "log" {
		return flat.OpenLog(file, c.Pattern)
	}
	return workbook.OpenFormat(file, c.Format)
}

func (c GetInfoCommand) printInfo(info grid.ViewInfo, j int) {
	var (
		active string
		state  = "visible"
		locked = "unlocked"
	)
	if info.Hidden {
		state = "hidden"
	}
	if info.Active {
		active = "*"
	} else {
		active = " "
	}
	if info.Protected {
		locked = "locked"
	}

	fmt.Fprintf(os.Stdout, infoPattern, j+1, active, info.Name, state, info.Size.Lines, info.Size.Columns, locked)
	fmt.Fprintln(os.Stdout)
}
