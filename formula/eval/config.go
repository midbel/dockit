package eval

import (
	"fmt"
	"os"

	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/internal/slx"
)

var (
	ConfigContextDir      = slx.Make("context", "dir")
	ConfigPrintDebug      = slx.Make("print", "debug")
	ConfigPrintCols       = slx.Make("print", "cols")
	ConfigPrintRows       = slx.Make("print", "rows")
	ConfigFormatNumber    = slx.Make("format", "number")
	ConfigFormatDate      = slx.Make("format", "date")
	ConfigFormatBool      = slx.Make("format", "bool")
	ConfigImportCsvDelim  = slx.Make("import", "csv", "delimiter")
	ConfigImportCsvQuoted = slx.Make("import", "csv", "quoted")
	ConfigAssertMode      = slx.Make("assert", "mode")
	ConfigExportFormat    = slx.Make("export", "format")
	ConfigCopyMode        = slx.Make("copy", "mode")
)

var defaultConfig = []struct {
	Key   []string
	Value any
}{
	{
		Key:   ConfigContextDir,
		Value: ".",
	},
	{
		Key:   ConfigPrintDebug,
		Value: false,
	},
	{
		Key:   ConfigPrintCols,
		Value: float64(maxCols),
	},
	{
		Key:   ConfigPrintRows,
		Value: float64(maxRows),
	},
	{
		Key:   ConfigFormatNumber,
		Value: format.DefaultNumberPattern,
	},
	{
		Key:   ConfigFormatDate,
		Value: format.DefaultDatePattern,
	},
	{
		Key:   ConfigFormatBool,
		Value: "",
	},
	{
		Key:   ConfigImportCsvDelim,
		Value: "comma",
	},
	{
		Key:   ConfigImportCsvQuoted,
		Value: true,
	},
	{
		Key:   ConfigAssertMode,
		Value: "fail",
	},
	{
		Key:   ConfigExportFormat,
		Value: "oxml",
	},
	{
		Key:   ConfigCopyMode,
		Value: false,
	},
}

type EngineConfig struct {
	registry *ds.Trie[any]
}

func NewConfig() *EngineConfig {
	c := EngineConfig{
		registry: ds.NewTrie[any](),
	}
	c.SetDefaults()
	return &c
}

func (c *EngineConfig) ContextDir() string {
	dir, _ := c.registry.Get(ConfigContextDir)
	return dir.(string)
}

func (c *EngineConfig) Printer() (Printer, error) {
	debug, _ := c.registry.Get(ConfigPrintDebug)
	cols, _ := c.registry.Get(ConfigPrintCols)
	rows, _ := c.registry.Get(ConfigPrintRows)

	maxCols, ok := cols.(float64)
	if !ok {
		return nil, fmt.Errorf("columns should be a number")
	}
	maxRows, ok := rows.(float64)
	if !ok {
		return nil, fmt.Errorf("rows should be a number")
	}
	if d, ok := debug.(bool); ok && d {
		return DebugValue(os.Stdout, int(maxRows), int(maxCols)), nil
	}
	return PrintValue(os.Stdout, int(maxRows), int(maxCols)), nil
}

func (c *EngineConfig) Formatter() (format.Formatter, error) {
	vf := format.FormatValue()
	if num, ok := c.registry.Get(ConfigFormatNumber); ok {
		str, ok := num.(string)
		if !ok {
			return nil, fmt.Errorf("number pattern should be a literal")
		}
		if err := vf.Number(str); err != nil {
			return nil, err
		}
	}
	if date, ok := c.registry.Get(ConfigFormatDate); ok {
		str, ok := date.(string)
		if !ok {
			return nil, fmt.Errorf("date pattern should be a literal")
		}
		if err := vf.Date(str); err != nil {
			return nil, err
		}
	}
	if mode, ok := c.registry.Get(ConfigFormatBool); ok {
		str, ok := mode.(string)
		if !ok {
			return nil, fmt.Errorf("boolean pattern should be a literal")
		}
		if err := vf.Bool(str); err != nil {
			return nil, err
		}
	}
	return vf, nil
}

func (c *EngineConfig) Set(ident []string, val any) error {
	if val == nil {
		return nil
	}
	c.registry.Register(ident, val)
	return nil
}

func (c *EngineConfig) Get(key []string) any {
	v, _ := c.registry.Get(key)
	return v
}

func (c *EngineConfig) SetDefaults() {
	for _, v := range defaultConfig {
		c.Set(v.Key, v.Value)
	}
}

func (c *EngineConfig) Merge(other *EngineConfig) error {
	return c.registry.Merge(other.registry)
}

func csvDelimiter(value string) string {
	switch value {
	default:
		return value
	case "pipe":
		return "|"
	case "comma", "":
		return ","
	case "semi", "semicolon":
		return ";"
	case "tab":
		return "\t"
	case "space":
		return " "
	}
}

func assertMode(value string) parse.AssertType {
	switch value {
	default:
		return parse.AssertFail
	case "warn":
		return parse.AssertWarn
	case "ignore":
		return parse.AssertIgnore
	}
}
