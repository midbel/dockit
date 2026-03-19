package format

import (
	"fmt"
	"strconv"

	"github.com/midbel/dockit/internal/ds"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var PatternNames = map[string]string{
	"ISO": "YYYY-MM-DD",
}

const (
	DefaultNumberPattern = "#######.00"
	DefaultDatePattern   = "YYYY-MM-DD"
	DefaultBoolPattern   = "bool"
)

type Formatter interface {
	Format(value.Value) (string, error)
}

type formattedValue struct {
	inner     value.Value
	formatter Formatter
}

func (v formattedValue) Kind() value.ValueKind {
	return v.inner.Kind()
}

func (v formattedValue) Type() string {
	return v.inner.Type()
}

func (v formattedValue) String() string {
	return v.inner.String()
}

func (v formattedValue) Format() (string, error) {
	return v.formatter.Format(v.inner)
}

type ValueFormatter struct {
	registry *ds.Trie[Formatter]
}

func FormatValue() *ValueFormatter {
	vf := ValueFormatter{
		registry: ds.NewTrie[Formatter](),
	}
	return &vf
}

func (vf *ValueFormatter) Set(kind string, formatter Formatter) {
	vf.registry.Register(slx.One(kind), formatter)
}

func (vf *ValueFormatter) Number(pattern string) error {
	f, err := ParseNumberFormatter(pattern)
	if err == nil {
		vf.Set(value.TypeNumber, f)
	}
	return err
}

func (vf *ValueFormatter) Date(pattern string) error {
	f, err := ParseDateFormatter(pattern)
	if err == nil {
		vf.Set(value.TypeDate, f)
	}
	return err
}

func (vf *ValueFormatter) Bool(pattern string) error {
	var mode boolMode
	switch pattern {
	case "", "bool":
		mode = boolDefault
	case "yesno":
		mode = boolYesNo
	case "onoff":
		mode = boolOnOff
	default:
		return fmt.Errorf("unknown boolean pattern")
	}
	f := boolFormatter{
		mode: mode,
	}
	vf.Set(value.TypeBool, f)
	return nil
}

func (vf *ValueFormatter) Format(v value.Value) (string, error) {
	if v, ok := v.(formattedValue); ok {
		return v.Format()
	}
	f, ok := vf.registry.Get(slx.One(v.Type()))
	if ok {
		return f.Format(v)
	}
	return v.String(), nil
}

type strFormatter struct{}

func FormatString() Formatter {
	var s strFormatter
	return s
}

func (strFormatter) Format(v value.Value) (string, error) {
	return v.String(), nil
}

type boolMode int

const (
	boolDefault boolMode = iota
	boolYesNo
	boolOnOff
)

type boolFormatter struct {
	mode boolMode
}

func FormatBool() Formatter {
	var b boolFormatter
	b.mode = boolDefault
	return b
}

func FormatYesNo() Formatter {
	var b boolFormatter
	b.mode = boolYesNo
	return b
}

func FormatOnOff() Formatter {
	var b boolFormatter
	b.mode = boolOnOff
	return b
}

func (f boolFormatter) Format(v value.Value) (string, error) {
	b, ok := v.(value.Boolean)
	if !ok {
		return "", fmt.Errorf("value is not a boolean")
	}
	switch f.mode {
	case boolDefault:
		return strconv.FormatBool(bool(b)), nil
	case boolYesNo:
		if b {
			return "yes", nil
		}
		return "no", nil
	case boolOnOff:
		if b {
			return "on", nil
		}
		return "off", nil
	default:
		return "", fmt.Errorf("invalid mode")
	}
}
