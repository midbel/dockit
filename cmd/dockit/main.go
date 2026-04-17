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
	summary = "Dockit transforms the way you handle spreadsheets by moving manual data tasks into your terminal"
	help    = `Dockit CLI is a data processing tool designed to manipulate, transform, and export tabular data directly from your terminal. 

Built to bridge the gap between raw data and spreadsheets, with Dockit CLI, you can:

* Manage Sheets: Seamlessly add, remove, or reorganize sheets within your workbooks
* Unified Format Support: Interface with .csv, .xlsx, and .ods files using a single, consistent toolset
* Restructure Data: Transform and reshape sheet layouts to fit your specific requirements
* Join & Merge: Combine sheets across the same workbook or consolidate data from multiple different files

Finally, with Dockit, manipulating spreadsheets from the command line becomes a workflow.`
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
	root.SetSummary(summary)
	root.SetHelp(help)

	root.Register(slx.One("info"), &infoCmd)
	root.Register(slx.One("merge"), &mergeCmd)
	root.Register(slx.One("format"), &formatCmd)
	root.Register(slx.One("run"), &runCmd)
	root.Register(slx.One("dump"), &dumpCmd)
	root.Register(slx.One("lock"), &lockCmd)
	root.Register(slx.One("unlock"), &unlockCmd)
	root.Register(slx.One("add"), &addCmd)
	root.Register(slx.One("join"), &joinCmd)
	root.Register(slx.One("group"), &groupCmd)
	root.Register(slx.One("transpose"), &transposeCmd)
	root.Register(slx.One("drop"), &dropCmd)
	root.Register(slx.One("rename"), &renameCmd)
	root.Register(slx.One("copy"), &copyCmd)
	root.Register(slx.One("print"), &printCmd)
	// root.Register(slx.One("deps"), &depsCmd)
	root.Register(slx.One("audit"), &auditCmd)
	root.Register(slx.One("builtins"), &builtinsCmd)

	// root.Register(slx.Make("studio", "browse-ast"), &terminalBrowseAstCmd)

	return root
}

func withSheet(path, name string, fn func(grid.View) error) error {
	wb, err := workbook.Open(path)
	if err != nil {
		return err
	}
	var sh grid.View
	if name == "" {
		sh, err = wb.ActiveSheet()
	} else {
		sh, err = wb.Sheet(name)
	}
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
