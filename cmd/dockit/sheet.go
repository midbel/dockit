package main

import (
	"github.com/midbel/cli"
	"github.com/midbel/dockit/internal/slx"
)

var addCmd = cli.Command{
	Name:    "add",
	Alias:   slx.Make("append"),
	Summary: "add one or multiple sheets from a spreadsheet like file to another",
	Usage:   "",
	Handler: &AddCommand{},
}

var dropCmd = cli.Command{
	Name:    "drop",
	Alias:   slx.Make("remove", "rm"),
	Summary: "remove one or multiple sheets from a spreadsheet file",
	Usage:   "",
	Handler: &DropCommand{},
}

var renameCmd = cli.Command{
	Name:    "rename",
	Summary: "rename a sheet from a spreadsheet file",
	Usage:   "",
	Handler: &RenameCommand{},
}

var copyCmd = cli.Command{
	Name:    "copy",
	Alias:   slx.Make("cp"),
	Summary: "copy a sheet from a spreadsheet file to the same file to another",
	Usage:   "",
	Handler: &CopyCommand{},
}

var moveCmd = cli.Command{
	Name:    "move",
	Alias:   slx.Make("mv"),
	Summary: "move a sheet from a spreadsheet file to the same file to another",
	Usage:   "",
	Handler: &MoveCommand{},
}

type AddCommand struct{}

func (c AddCommand) Run(args []string) error {
	return nil
}

type DropCommand struct{}

func (c DropCommand) Run(args []string) error {
	return nil
}

type CopyCommand struct{}

func (c CopyCommand) Run(args []string) error {
	return nil
}

type MoveCommand struct{}

func (c MoveCommand) Run(args []string) error {
	return nil
}

type RenameCommand struct{}

func (c RenameCommand) Run(args []string) error {
	return nil
}
