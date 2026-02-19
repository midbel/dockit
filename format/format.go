package format

import (
	"github.com/midbel/dockit/value"
)

const (
	DefaultNumberPattern = "#######.00"
	DefaultDatePattern   = "YYYY-MM-DD"
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
	formatters map[string]Formatter
}

func FormatValue() *ValueFormatter {
	vf := ValueFormatter{
		formatters: make(map[string]Formatter),
	}
	return &vf
}

func (vf *ValueFormatter) Set(kind string, formatter Formatter) {
	vf.formatters[kind] = formatter
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

func (vf *ValueFormatter) Format(v value.Value) (string, error) {
	f, ok := vf.formatters[v.Type()]
	if ok {
		return f.Format(v)
	}
	return v.String(), nil
}

type strFormatter struct{}

func (strFormatter) Format(v value.Value) (string, error) {
	return v.String(), nil
}

type boolFormatter struct{}

func (f boolFormatter) Format(v value.Value) (string, error) {
	return "", nil
}
