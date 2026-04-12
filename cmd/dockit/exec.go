package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/eval"
	"github.com/midbel/dockit/formula/repr"
)

var runCmd = cli.Command{
	Name:    "run",
	Summary: "Execute given script",
	Usage:   "run [-g] [-d <dir>] <script.dk>",
	Handler: &RunCommand{},
}

var dumpCmd = cli.Command{
	Name:    "dump",
	Alias:   []string{"inspect"},
	Summary: "Export the AST representation of script",
	Usage:   "dump <script.dk>",
	Handler: &DumpCommand{},
}

type RunCommand struct {
	Debug        bool
	Dialect      string
	ContextDir   string
	DateFormat   string
	NumberFormat string
}

func (c RunCommand) Run(args []string) error {
	set := cli.NewFlagSet("run")
	set.BoolVar(&c.Debug, "g", false, "print debug")
	set.StringVar(&c.ContextDir, "d", ".", "Context directory")
	if err := set.Parse(args); err != nil {
		return err
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		return err
	}
	defer r.Close()

	engine := eval.NewEngine()
	engine.SetPrintDebug(c.Debug)
	engine.SetContextDir(c.ContextDir)
	engine.SetNumberFormat(c.NumberFormat)
	engine.SetDateFormat(c.DateFormat)
	_, err = engine.Exec(r, env.Empty())
	return err
}

type DumpCommand struct{}

func (c DumpCommand) Run(args []string) error {
	set := cli.NewFlagSet("dump")
	if err := set.Parse(args); err != nil {
		return err
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		return err
	}
	defer r.Close()

	root, err := repr.Inspect(r)
	if err != nil {
		return err
	}
	printNode(os.Stdout, root.Root, 0)
	return nil
}

func printNode(w io.Writer, node *repr.Node, level int) {
	prefix := strings.Repeat(" ", level*2)
	io.WriteString(w, prefix)
	io.WriteString(w, node.Type)
	io.WriteString(w, "[")
	io.WriteString(w, node.Name)
	io.WriteString(w, "]")
	if node.Value == nil && len(node.Params) == 0 && len(node.Children) == 0 {
		io.WriteString(w, "\n")
		return
	}
	io.WriteString(w, "[\n")
	if node.Value != nil {
		io.WriteString(w, prefix)
		io.WriteString(w, "  value: ")
		io.WriteString(w, fmt.Sprint(node.Value))
		io.WriteString(w, "\n")
	}
	for _, p := range node.Params {
		if p.Name == "value" && node.Value != nil {
			continue
		}
		io.WriteString(w, prefix)
		io.WriteString(w, "  ")
		io.WriteString(w, p.Name)
		io.WriteString(w, ": ")
		io.WriteString(w, fmt.Sprint(p.Value))
		io.WriteString(w, "\n")
	}
	for _, n := range node.Children {
		printNode(w, n, level+1)
	}
	io.WriteString(w, prefix)
	io.WriteString(w, "]\n")
}
