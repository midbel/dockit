package formula

import (
	"strconv"

	"github.com/midbel/dockit/layout"
	"github.com/midbel/dockit/value"
)

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

type Blank struct{}

func (Blank) Kind() value.ValueKind {
	return value.KindScalar
}

func (Blank) String() string {
	return ""
}

func (Blank) Scalar() any {
	return nil
}

type Float float64

func (Float) Kind() value.ValueKind {
	return value.KindScalar
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Scalar() any {
	return float64(f)
}

type Text string

func (Text) Kind() value.ValueKind {
	return value.KindScalar
}

func (t Text) String() string {
	return string(t)
}

func (t Text) Scalar() any {
	return string(t)
}

type Boolean bool

func (Boolean) Kind() value.ValueKind {
	return value.KindScalar
}

func (b Boolean) String() string {
	return strconv.FormatBool(bool(b))
}

func (b Boolean) Scalar() any {
	return bool(b)
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
