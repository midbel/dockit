package main

import (
	"github.com/midbel/cli"

	tea "charm.land/bubbletea/v2"
	"github.com/midbel/dockit/terminal/ast"
)

var terminalBrowseAstCmd = cli.Command{
	Name:    "view-ast",
	Summary: "",
	Usage:   "",
	Handler: &BrowseAstCommand{},
}

type BrowseAstCommand struct{}

func (c BrowseAstCommand) Run(args []string) error {
	set := cli.NewFlagSet("browse-ast")
	if err := set.Parse(args); err != nil {
		return err
	}
	p := tea.NewProgram(ast.NewModel(set.Arg(0)))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
