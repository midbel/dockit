package types

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

func IsComparable(v value.Value) bool {
	_, ok := v.(value.Comparable)
	return ok
}

func IsNumber(v value.Value) bool {
	_, ok := v.(Float)
	return ok
}

func IsScalar(v value.Value) bool {
	return v.Kind() == value.KindScalar
}

var ErrCompatible = errors.New("incompatible type")

type ErrorCode string

var (
	ErrNull  = createError("#NULL!")
	ErrDiv0  = createError("#DIV/0!")
	ErrValue = createError("#VALUE!")
	ErrRef   = createError("#REF!")
	ErrName  = createError("#NAME?")
	ErrNum   = createError("#NUM!")
	ErrNA    = createError("#N/A")
)

type Error struct {
	code string
}

func createError(code string) Error {
	return Error{
		code: code,
	}
}

func (Error) Kind() value.ValueKind {
	return value.KindError
}

func (e Error) Error() string {
	return e.code
}

func (e Error) String() string {
	return e.code
}

func (e Error) Scalar() any {
	return e.code
}

type Array struct {
	Data [][]value.ScalarValue
}

func (Array) Kind() value.ValueKind {
	return value.KindArray
}

func (Array) String() string {
	return ""
}

func (a Array) Dimension() layout.Dimension {
	var (
		d layout.Dimension
		n = len(a.Data)
	)
	if n > 0 {
		d.Lines = int64(n)
		d.Columns = int64(len(a.Data[0]))
	}
	return d
}

func (a Array) At(row, col int) value.ScalarValue {
	if len(a.Data) == 0 || row >= len(a.Data) {
		return nil
	}
	v := a.Data[row]
	if len(v) == 0 || col >= len(v) {
		return nil
	}
	return a.Data[row][col]
}
