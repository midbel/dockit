package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/doc"
	_ "github.com/midbel/dockit/formula/eval"
	"github.com/midbel/dockit/grid"
)

var errFail = errors.New("fail")

var (
	summary = "dockit"
	help    = ""
)

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
	root.Register([]string{"info"}, &infoCmd)

	return root
}

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "get informations about sheets in given file",
	Usage:   "info [-a] <spreadsheet> [<sheet>,...]",
	Handler: &GetInfoCommand{},
}

const infoPattern = "%d %s%s(%s): %d lines, %d columns - %s"

type GetInfoCommand struct {
	Format doc.Format
}

func (c GetInfoCommand) Run(args []string) error {
	set := cli.NewFlagSet("info")
	set.Func("f", "format", func(str string) error {
		return nil
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	infos, err := doc.Infos(set.Arg(0))
	if err != nil {
		return err
	}
	for j, i := range infos {
		c.printInfo(i, j)
	}
	return nil
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
