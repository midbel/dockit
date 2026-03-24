package criteria

import (
	"strings"

	"github.com/midbel/dockit/value"
)

type Filter interface {
	Keep(value.Value) bool
}

func New(predicate string) (Filter, error) {
	predicate = strings.TrimSpace(predicate)

	var f valFilter
	switch {
	case strings.HasPrefix(predicate, "="):
		f.value = value.Text(predicate[1:])
		f.cmp = value.Eq
	case strings.HasPrefix(predicate, "<>"):
		f.value = value.Text(predicate[2:])
		f.cmp = value.Ne
	case strings.HasPrefix(predicate, ">"):
		f.value = value.Text(predicate[1:])
		f.cmp = value.Gt
	case strings.HasPrefix(predicate, ">="):
		f.value = value.Text(predicate[2:])
		f.cmp = value.Ge
	case strings.HasPrefix(predicate, "<"):
		f.value = value.Text(predicate[1:])
		f.cmp = value.Lt
	case strings.HasPrefix(predicate, "<="):
		f.value = value.Text(predicate[2:])
		f.cmp = value.Le
	default:
		f.value = value.Text(predicate)
		f.cmp = value.Eq
	}
	return f, nil
}

func Match(val value.Value, predicate string) bool {
	f, err := New(predicate)
	if err != nil {
		return false
	}
	return f.Keep(val)
}

type valFilter struct {
	value value.Value
	cmp   func(value.Value, value.Value) value.Value
}

func (f valFilter) Keep(val value.Value) bool {
	var (
		other value.Value
		err   error
	)
	switch val.Type() {
	case value.TypeNumber:
		other, err = value.CastToFloat(f.value)
	case value.TypeText:
		other, err = value.CastToText(f.value)
	case value.TypeDate:
		other, err = value.CastToDate(f.value)
	case value.TypeBool:
		t := value.True(f.value)
		other = value.Boolean(t)
	default:
		return false
	}
	if err != nil {
		return false
	}
	res := f.cmp(val, other)
	return value.True(res)
}
