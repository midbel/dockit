package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/oxml"
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
	root.Register([]string{"lookup"}, &lookupCmd)
	root.Register([]string{"merge"}, &mergeCmd)
	root.Register([]string{"append"}, &appendCmd)
	root.Register([]string{"move"}, &moveCmd)
	root.Register([]string{"copy"}, &copyCmd)
	root.Register([]string{"extract"}, &extractCmd)
	root.Register([]string{"join"}, &joinCmd)

	return root
}

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "get informations about sheets in given file",
	Handler: &GetInfoCommand{},
}

var lookupCmd = cli.Command{
	Name:    "lookup",
	Alias:   []string{"search", "find"},
	Summary: "find that data in given file",
	Handler: nil,
}

var mergeCmd = cli.Command{
	Name:    "merge",
	Summary: "merge sheets of file(s) into a single file",
	Handler: &MergeFilesCommand{},
}

var appendCmd = cli.Command{
	Name:    "append",
	Summary: "append data from input to file to a sheet in a given file",
	Handler: nil,
}

var joinCmd = cli.Command{
	Name:    "join",
	Summary: "join sheets of a file",
	Handler: nil,
}

var extractCmd = cli.Command{
	Name:    "extract",
	Summary: "extract one or more sheets from given file",
	Handler: &ExtractSheetCommand{},
}

var moveCmd = cli.Command{
	Name:    "move",
	Alias:   []string{"mv"},
	Summary: "move one or more sheets from one file to another",
	Handler: nil,
}

var copyCmd = cli.Command{
	Name:    "copy",
	Alias:   []string{"cp"},
	Summary: "move one or more sheets from one file to another",
	Handler: nil,
}

var convertCmd = cli.Command{
	Name:    "convert",
	Summary: "convert file to another format",
	Handler: nil,
}

type ExtractSheetCommand struct {
	OutDir    string
	Format    string
	Delimiter string
	Range     string
}

func (c ExtractSheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("extract")
	set.StringVar(&c.OutDir, "d", "", "write result to output file")
	set.StringVar(&c.Format, "f", "", "extract to given format (csv, json, xml)")
	set.StringVar(&c.Delimiter, "c", "", "delimiter to use")
	set.StringVar(&c.Range, "r", "", "range of data to extract")
	if err := set.Parse(args); err != nil {
		return err
	}
	return nil
}

type MergeFilesCommand struct {
	OutFile string
}

func (c MergeFilesCommand) Run(args []string) error {
	set := cli.NewFlagSet("merge")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	if err := set.Parse(args); err != nil {
		return err
	}
	f, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	for i := 1; i < set.NArg(); i++ {
		isZip, err := isZip(set.Arg(i))
		if err != nil {
			return err
		}
		if isZip {
			err = mergeFile(f, set.Arg(i))
		} else {
			err = mergeCSV(f, set.Arg(i))
		}
		if err != nil {
			return err
		}
	}
	return f.WriteFile(c.OutFile)
}

func mergeCSV(f *oxml.File, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	name := filepath.Base(file)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	sheet := oxml.NewSheet(name)

	rs := csv.NewReader(r)
	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		sheet.Append(row)
	}
	return f.AppendSheet(sheet)
}

func mergeFile(f *oxml.File, file string) error {
	x, err := oxml.Open(file)
	if err != nil {
		return err
	}
	return f.Merge(x)
}

var magicZipBytes = [][]byte{
	{0x50, 0x4b, 0x03, 0x04},
	{0x50, 0x4b, 0x05, 0x06},
	{0x50, 0x4b, 0x07, 0x08},
}

func isZip(file string) (bool, error) {
	r, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer r.Close()

	magic := make([]byte, 4)
	if n, err := io.ReadFull(r, magic); err != nil || n != len(magic) {
		return false, err
	}
	for _, mzb := range magicZipBytes {
		if bytes.Equal(magic, mzb) {
			return true, nil
		}
	}
	return false, nil
}

type GetInfoCommand struct{}

func (c GetInfoCommand) Run(args []string) error {
	set := cli.NewFlagSet("info")
	if err := set.Parse(args); err != nil {
		return err
	}
	f, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	var (
		sheets  = f.Sheets()
		pattern = "%d %s: %d lines, %d columns"
	)
	for _, s := range sheets {
		fmt.Fprintf(os.Stdout, pattern, s.Index, s.Name, s.Size.Lines, s.Size.Columns)
		fmt.Fprintln(os.Stdout)
	}
	return nil
}
