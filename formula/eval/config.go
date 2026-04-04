package eval

import (
	"fmt"
	"os"

	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/internal/slx"
)

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
	dir, _ := c.registry.Get(slx.Make("context", "dir"))
	return dir.(string)
}

func (c *EngineConfig) Printer() (Printer, error) {
	debug, _ := c.registry.Get(slx.Make("print", "debug"))
	cols, _ := c.registry.Get(slx.Make("print", "cols"))
	rows, _ := c.registry.Get(slx.Make("print", "rows"))

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
	if num, ok := c.registry.Get(slx.Make("format", "number")); ok {
		str, ok := num.(string)
		if !ok {
			return nil, fmt.Errorf("number pattern should be a literal")
		}
		if err := vf.Number(str); err != nil {
			return nil, err
		}
	}
	if date, ok := c.registry.Get(slx.Make("format", "date")); ok {
		str, ok := date.(string)
		if !ok {
			return nil, fmt.Errorf("date pattern should be a literal")
		}
		if err := vf.Date(str); err != nil {
			return nil, err
		}
	}
	if mode, ok := c.registry.Get(slx.Make("format", "bool")); ok {
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

func (c *EngineConfig) SetDefaults() {
	c.Set(slx.Make("context", "dir"), ".")
	c.Set(slx.Make("print", "debug"), false)
	c.Set(slx.Make("print", "cols"), float64(maxCols))
	c.Set(slx.Make("print", "rows"), float64(maxRows))
	c.Set(slx.Make("format", "number"), format.DefaultNumberPattern)
	c.Set(slx.Make("format", "date"), format.DefaultDatePattern)
	c.Set(slx.Make("format", "bool"), "")
	c.Set(slx.Make("csv", "delimiter"), ",")
	c.Set(slx.Make("csv", "quoted"), true)
}

func (c *EngineConfig) Merge(other *EngineConfig) error {
	return c.registry.Merge(other.registry)
}
