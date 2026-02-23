package main

type RunCommand struct {}

func (c RunCommand) Run(args []string) error {
	return nil
}

type DumpCommand struct{}

func (c DumpCommand) Run(args []string) error {
	return nil
}