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
	"github.com/midbel/dockit/ods"
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
	workbook.Register(ods.NewLoader())
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
	root.Register(slx.One("lock"), &lockCmd)
	root.Register(slx.One("unlock"), &unlockCmd)
	root.Register(slx.One("add"), &addCmd)
	root.Register(slx.One("drop"), &dropCmd)
	root.Register(slx.One("rename"), &renameCmd)
	root.Register(slx.One("copy"), &copyCmd)
	root.Register(slx.One("print"), &printCmd)
	root.Register(slx.One("deps"), &depsCmd)
	root.Register(slx.One("audit"), &auditCmd)
	root.Register(slx.One("builtins"), &builtinsCmd)

	root.Register(slx.Make("studio", "browse-ast"), &terminalBrowseAstCmd)

	return root
}

func withSheet(path, name string, fn func(grid.View) error) error {
	wb, err := workbook.Open(path)
	if err != nil {
		return err
	}
	sh, err := wb.Sheet(name)
	if err == nil {
		err = fn(sh)
	}
	return err
}

func updateFile(path string, fn func(grid.File) error) error {
	wb, err := workbook.Open(path)
	if err != nil {
		return err
	}
	if err := fn(wb); err != nil {
		return err
	}
	return workbook.WriteFile(wb, path)
}
