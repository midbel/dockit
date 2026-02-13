package types

import (
	"fmt"
	"slices"

	"github.com/midbel/dockit/value"
)

const (
	InspectKindCell  = "cell"
	InspectKindView  = "view"
	InspectKindSheet = "sheet"
	InspectKindFile  = "file"
	InspectKindRange = "range"
	InspectKindSlice = "slice"
	InspectKindValue = "value" // fallback for computed expressions
)

type InspectField struct {
	Name  string
	Value value.Value
}

type InspectValue struct {
	source *InspectValue
	fields []*InspectField
}

func InspectCell() *InspectValue {
	return inspectValue(InspectKindCell)
}

func InspectView() *InspectValue {
	return inspectValue(InspectKindView)
}

func InspectRange() *InspectValue {
	return inspectValue(InspectKindRange)
}

func InspectSlice() *InspectValue {
	return inspectValue(InspectKindSlice)
}

func ReinspectValue(iv *InspectValue, val value.Value) *InspectValue {
	iv.Set("type", value.Text(val.Type()))
	iv.Set("value", val)
	return iv
}

func inspectValue(kind string) *InspectValue {
	var iv InspectValue
	iv.Set("kind", value.Text(kind))
	return &iv
}

func (v *InspectValue) SetSource(iv *InspectValue) {
	v.source = iv
}

func (v *InspectValue) Source() *InspectValue {
	return v.source
}

func (*InspectValue) Type() string {
	return "inspect"
}

func (*InspectValue) Kind() value.ValueKind {
	return value.KindInspectable
}

func (*InspectValue) String() string {
	return "<inspect>"
}

func (v *InspectValue) Set(name string, val value.Value) {
	f := InspectField{
		Name:  name,
		Value: val,
	}
	v.fields = append(v.fields, &f)
}

func (v *InspectValue) Get(ident string) (value.Value, error) {
	if ident == "source" {
		return v.source, nil
	}
	ix := slices.IndexFunc(v.fields, func(f *InspectField) bool {
		return f.Name == ident
	})
	if ix < 0 {
		return nil, fmt.Errorf("%s: %w", ident, value.ErrProp)
	}
	return v.fields[ix].Value, nil
}
