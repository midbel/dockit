package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/csv"
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
	root.Register([]string{"print"}, &printCmd)
	root.Register([]string{"lookup"}, &lookupCmd)
	root.Register([]string{"new"}, &newCmd)
	root.Register([]string{"merge"}, &mergeCmd)
	root.Register([]string{"append"}, &appendCmd)
	root.Register([]string{"move"}, &moveCmd)
	root.Register([]string{"copy"}, &copyCmd)
	root.Register([]string{"remove"}, &removeCmd)
	root.Register([]string{"extract"}, &extractCmd)
	root.Register([]string{"join"}, &joinCmd)
	root.Register([]string{"eval"}, &evalCmd)
	root.Register([]string{"split"}, &splitCmd)
	root.Register([]string{"transpose"}, &transposeCmd)
	root.Register([]string{"lock"}, &lockCmd)
	root.Register([]string{"unlock"}, &unlockCmd)

	return root
}

var infoCmd = cli.Command{
	Name:    "info",
	Summary: "get informations about sheets in given file",
	Usage:   "info [-a] <spreadsheet> [<sheet>,...]",
	Handler: &GetInfoCommand{},
}

var lookupCmd = cli.Command{
	Name:    "lookup",
	Alias:   []string{"search", "find"},
	Summary: "find that data in given file",
	Handler: nil,
}

var printCmd = cli.Command{
	Name:    "print",
	Alias:   []string{"view", "show", "dump"},
	Summary: "print content of a sheet",
	Usage:   "print [-c <columns>] [-r] <spreadsheet> [<sheet>,...]",
	Handler: &PrintSheetCommand{},
}

var newCmd = cli.Command{
	Name:    "new",
	Alias:   []string{"create"},
	Summary: "create a new spreadsheet from input files",
	Usage:   "new [-o file] <file, [file,...]>",
	Handler: &CreateFileCommand{},
}

var mergeCmd = cli.Command{
	Name:    "merge",
	Summary: "merge sheets of one or more spreadsheet into one",
	Usage:   "merge [-o file] <spreadsheet> <file1,[file2,...]>",
	Handler: &MergeFilesCommand{},
}

var appendCmd = cli.Command{
	Name:    "append",
	Summary: "append data to a spreadsheet",
	Handler: nil,
}

var joinCmd = cli.Command{
	Name:    "join",
	Summary: "join multiple sheets of a spreadsheet",
	Handler: nil,
}

var evalCmd = cli.Command{
	Name:    "eval",
	Summary: "execute script on spreadsheet",
	Handler: nil,
}

var splitCmd = cli.Command{
	Name:    "split",
	Summary: "split a sheet to multiple according to given criteria",
	Handler: nil,
}

var transposeCmd = cli.Command{
	Name:    "transpose",
	Summary: "switch rows and columns",
	Handler: nil,
}

var extractCmd = cli.Command{
	Name:    "extract",
	Alias:   []string{"export"},
	Summary: "extract one or more sheets from given spreadsheets",
	Usage:   "extract [-d directory] [-f format] [-c delimiter] <spreadsheet> [sheet,...]",
	Handler: &ExtractSheetCommand{},
}

var moveCmd = cli.Command{
	Name:    "move",
	Alias:   []string{"mv"},
	Summary: "rename or move a sheet from one spreadsheet to another",
	Usage:   "move file:sheet[:target] [file[:sheet]]",
	Handler: &MoveSheetCommand{},
}

var copyCmd = cli.Command{
	Name:    "copy",
	Alias:   []string{"cp"},
	Summary: "copy a sheet from one spreadsheet to the same or another",
	Usage:   "copy [-o] [-c] file:sheet[:target] [file[:sheet]]",
	Handler: &CopySheetCommand{},
}

var removeCmd = cli.Command{
	Name:    "remove",
	Alias:   []string{"rm"},
	Summary: "remove a sheet from its spreadfile",
	Usage:   "remove file:sheet",
	Handler: &RemoveSheetCommand{},
}

var convertCmd = cli.Command{
	Name:    "convert",
	Summary: "convert spreadsheet to another format",
	Handler: nil,
}

var lockCmd = cli.Command{
	Name:    "lock",
	Summary: "lock an entire spreadsheet or some of its sheet(s)",
	Handler: &LockFileCommand{},
}

var unlockCmd = cli.Command{
	Name:    "unlock",
	Summary: "unlock an entire spreadsheet or some of its sheet(s)",
	Handler: &UnlockFileCommand{},
}

type SheetRef struct {
	*oxml.File
	Path  string
	Sheet string
}

func parseReference(str string) (*SheetRef, error) {
	file, rest, ok := strings.Cut(str, ":")
	if !ok {
		return nil, fmt.Errorf("invalid reference given")
	}
	var (
		ref SheetRef
		err error
	)
	if ref.File, err = oxml.Open(file); err != nil {
		return nil, err
	}
	ref.Path = file
	ref.Sheet = rest
	if _, _, ok := strings.Cut(rest, "!"); ok {

	}
	return &ref, nil
}

func (s SheetRef) DumpTo(encoder oxml.Encoder, reload bool) error {
	sh, err := s.getSheetFromFile()
	if err != nil {
		return err
	}
	ctx := oxml.FileContext(s.File)
	if reload {
		if err := sh.Reload(ctx); err != nil {
			return err
		}
	}
	return sh.Encode(encoder)
}

func (s SheetRef) Remove(to string) error {
	if err := s.Remove(s.Sheet); err != nil {
		return err
	}
	if to == "" {
		to = s.Path
	}
	return s.WriteFile(to)
}

func (s SheetRef) Copy(to string) error {
	source, target, _ := strings.Cut(s.Sheet, ":")
	if err := s.File.Copy(source, target); err != nil {
		return err
	}
	if to == "" {
		to = s.Path
	}
	return s.WriteFile(to)
}

func (s SheetRef) CopyTo(other *SheetRef, to string) error {
	sh, err := s.getSheetFromFile()
	if err != nil {
		return err
	}
	if err := other.appendSheetToFile(sh); err != nil {
		return err
	}
	if to == "" {
		to = other.Path
	}
	return other.WriteFile(to)
}

func (s SheetRef) Move(to string) error {
	source, target, ok := strings.Cut(s.Sheet, ":")
	if !ok {
		return fmt.Errorf("missing new name")
	}
	if err := s.Rename(source, target); err != nil {
		return err
	}
	if to == "" {
		to = s.Path
	}
	return s.WriteFile(to)
}

func (s SheetRef) MoveTo(other *SheetRef, to string) error {
	sh, err := s.getSheetFromFile()
	if err != nil {
		return err
	}
	if err := other.appendSheetToFile(sh); err != nil {
		return err
	}
	if err := other.WriteFile(to); err != nil {
		return err
	}
	s.Remove(s.Sheet)
	return s.WriteFile(s.Path)
}

func (s SheetRef) appendSheetToFile(sh *oxml.Sheet) error {
	if s.Sheet != "" {
		sh.Name = s.Sheet
	}
	return s.AppendSheet(sh)
}

func (s SheetRef) getSheetFromFile() (*oxml.Sheet, error) {
	return s.File.Sheet(s.Sheet)
}

type RemoveSheetCommand struct {
	OutFile string
}

func (c RemoveSheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("remove")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	if err := set.Parse(args); err != nil {
		return err
	}
	ref, err := parseReference(set.Arg(0))
	if err != nil {
		return err
	}
	return ref.Remove(c.OutFile)
}

type CopySheetCommand struct {
	OutFile  string
	CopyOnly oxml.CopyMode
}

func (c CopySheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("copy")
	set.StringVar(&c.OutFile, "o", "", "write results to output file")
	set.Func("c", "", func(str string) error {
		mode, err := oxml.CopyModeFromString(str)
		if err == nil {
			c.CopyOnly = mode
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	switch set.NArg() {
	case 1:
		ref, err := parseReference(set.Arg(0))
		if err != nil {
			return err
		}
		return ref.Copy(c.OutFile)
	case 2:
		src, err := parseReference(set.Arg(0))
		if err != nil {
			return err
		}
		dst, err := parseReference(set.Arg(1))
		if err != nil {
			return err
		}
		return src.CopyTo(dst, c.OutFile)
	default:
		return fmt.Errorf("invalid number of arguments")
	}
	return nil
}

type MoveSheetCommand struct {
	OutFile string
}

func (c MoveSheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("move")
	if err := set.Parse(args); err != nil {
		return err
	}
	switch set.NArg() {
	case 1:
		ref, err := parseReference(set.Arg(0))
		if err != nil {
			return err
		}
		return ref.Move(c.OutFile)
	case 2:
		src, err := parseReference(set.Arg(0))
		if err != nil {
			return err
		}
		dst, err := parseReference(set.Arg(1))
		if err != nil {
			return err
		}
		return src.MoveTo(dst, c.OutFile)
	default:
		return fmt.Errorf("invalid number of arguments")
	}
	return nil
}

type LockFileCommand struct {
	OutFile string
}

func (c LockFileCommand) Run(args []string) error {
	set := cli.NewFlagSet("lock")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	if err := set.Parse(args); err != nil {
		return err
	}
	f, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	for i := 1; i < set.NArg(); i++ {
		if err := f.LockSheet(set.Arg(i)); err != nil {
			return err
		}
	}
	if c.OutFile == "" {
		c.OutFile = set.Arg(0)
	}
	return f.WriteFile(c.OutFile)
}

type UnlockFileCommand struct {
	OutFile string
}

func (c UnlockFileCommand) Run(args []string) error {
	set := cli.NewFlagSet("unlock")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	if err := set.Parse(args); err != nil {
		return err
	}
	f, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	for i := 1; i < set.NArg(); i++ {
		if err := f.UnlockSheet(set.Arg(i)); err != nil {
			return err
		}
	}
	if c.OutFile == "" {
		c.OutFile = set.Arg(0)
	}
	return f.WriteFile(c.OutFile)
}

type PrintSheetCommand struct {
	Columns string
	Width   int
	Sep     string
	Reload  bool
	Lino    bool
}

func (c PrintSheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("print")
	set.StringVar(&c.Columns, "c", "", "columns")
	set.BoolVar(&c.Reload, "r", false, "reload")
	set.StringVar(&c.Sep, "s", "|", "column separator")
	set.IntVar(&c.Width, "w", 12, "column width")
	set.BoolVar(&c.Lino, "n", false, "print line number")
	if err := set.Parse(args); err != nil {
		return err
	}
	ref, err := parseReference(set.Arg(0))
	if err != nil {
		return err
	}
	return ref.DumpTo(c, c.Reload)
}

func (c PrintSheetCommand) EncodeSheet(sheet *oxml.Sheet) error {
	if c.Width <= 0 {
		c.Width = 16
	}
	var lino int
	for row := range sheet.Data() {
		lino++
		if c.Lino {
			fmt.Fprintf(os.Stdout, "%-5d ", lino)
			fmt.Fprint(os.Stdout, c.Sep)
		}
		for i, v := range row {
			if i > 0 {
				fmt.Fprint(os.Stdout, c.Sep)
			}
			fmt.Fprintf(os.Stdout, " %-*v ", c.Width, v)
		}
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

type CreateFileCommand struct {
	OutFile string
	Sep     string
}

func (c CreateFileCommand) Run(args []string) error {
	set := cli.NewFlagSet("new")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	set.StringVar(&c.Sep, "s", "", "fields separator")
	if err := set.Parse(args); err != nil {
		return err
	}
	if c.OutFile == "" {
		c.OutFile = "new.xlsx"
	}
	if err := os.MkdirAll(filepath.Dir(c.OutFile), 0755); err != nil {
		return err
	}
	file := oxml.NewFile()
	for _, a := range set.Args() {
		if ok, err := isZip(a); ok || err != nil {
			continue
		}
		if err := mergeCSV(file, a, c.Sep); err != nil {
			return err
		}
	}
	return file.WriteFile(c.OutFile)
}

type ExtractSheetCommand struct {
	OutDir    string
	Format    string
	Delimiter string
	Range     string
}

func (c ExtractSheetCommand) Run(args []string) error {
	set := cli.NewFlagSet("extract")
	set.StringVar(&c.OutDir, "d", "", "write result to directory")
	set.StringVar(&c.Format, "f", "", "extract to given format (csv, json, xml)")
	set.StringVar(&c.Delimiter, "c", "", "delimiter to use")
	set.StringVar(&c.Range, "r", "", "range of data to extract")
	if err := set.Parse(args); err != nil {
		return err
	}
	if err := os.MkdirAll(c.OutDir, 0755); err != nil {
		return err
	}
	file, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	for i := 1; i < set.NArg(); i++ {
		err := c.Extract(file, set.Arg(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c ExtractSheetCommand) Extract(file *oxml.File, name string) error {
	sh, err := file.Sheet(name)
	if err != nil {
		return err
	}

	var (
		encode func(io.Writer) oxml.Encoder
		ext    string
	)
	switch c.Format {
	case "", "csv":
		encode = oxml.EncodeCSV
	case "json":
		encode = oxml.EncodeJSON
	case "xml":
		encode = oxml.EncodeXML
	default:
		return fmt.Errorf("%s: unsupported format", c.Format)
	}
	if c.Format == "" {
		c.Format = "csv"
	}
	ext = fmt.Sprintf(".%s", c.Format)

	w, err := os.Create(filepath.Join(c.OutDir, sh.Name+ext))
	if err != nil {
		return err
	}
	defer w.Close()

	return sh.Encode(encode(w))
}

type MergeFilesCommand struct {
	OutFile string
	Sep     string
}

func (c MergeFilesCommand) Run(args []string) error {
	set := cli.NewFlagSet("merge")
	set.StringVar(&c.OutFile, "o", "", "write result to output file")
	set.StringVar(&c.Sep, "s", "", "csv field separator")
	if err := set.Parse(args); err != nil {
		return err
	}
	f, err := oxml.Open(set.Arg(0))
	if err != nil {
		return err
	}
	if c.OutFile == "" {
		c.OutFile = set.Arg(0)
	}
	for i := 1; i < set.NArg(); i++ {
		isZip, err := isZip(set.Arg(i))
		if err != nil {
			return err
		}
		if isZip {
			err = mergeFile(f, set.Arg(i))
		} else {
			err = mergeCSV(f, set.Arg(i), c.Sep)
		}
		if err != nil {
			return err
		}
	}
	return f.WriteFile(c.OutFile)
}

func mergeCSV(f *oxml.File, file, sep string) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	name := filepath.Base(file)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	sheet := oxml.NewSheet(name)

	rs := csv.NewReader(r)
	if rs.Comma, err = csvSeparator(sep); err != nil {
		return err
	}

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
		pattern = "%d %s%s(%s): %d lines, %d columns - %s"
	)
	for _, s := range sheets {
		var (
			active string
			state  = s.Status()
			locked = "unlocked"
		)
		if s.Active {
			active = "*"
		} else {
			active = " "
		}
		if s.IsLock() {
			locked = "locked"
		}

		fmt.Fprintf(os.Stdout, pattern, s.Index, active, s.Name, state, s.Size.Lines, s.Size.Columns, locked)
		fmt.Fprintln(os.Stdout)
	}
	return nil
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

func csvSeparator(str string) (byte, error) {
	var comma byte
	switch str {
	case "semi", "semicolon", ";":
		comma = ';'
	case "comma", ",", "":
		comma = ','
	case "tab", "\t":
		comma = '\t'
	case "colon", ":":
		comma = ':'
	default:
		return 0, fmt.Errorf("unsupported separator")
	}
	return comma, nil
}
