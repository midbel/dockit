package value

import (
	"math"
	"strconv"
	"time"
)

type Blank struct{}

func Empty() ScalarValue {
	return Blank{}
}

func (Blank) Type() string {
	return "blank"
}

func (Blank) Kind() ValueKind {
	return KindScalar
}

func (Blank) String() string {
	return ""
}

func (Blank) Scalar() any {
	return nil
}

type Date time.Time

func (Date) Type() string {
	return "date"
}

func (Date) Kind() ValueKind {
	return KindScalar
}

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) Scalar() any {
	return time.Time(d)
}

func (d Date) Add(other Value) (ScalarValue, error) {
	f, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	t := time.Time(d).Add(time.Duration(f) * time.Second)
	return Date(t), nil
}

func (d Date) Sub(other Value) (ScalarValue, error) {
	f, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	t := time.Time(d).Add(-time.Duration(f) * time.Second)
	return Date(t), nil
}

func (d Date) ToString() (ScalarValue, error) {
	return Text(d.String()), nil
}

func (d Date) ToBool() (ScalarValue, error) {
	return Boolean(!time.Time(d).IsZero()), nil
}

func (d Date) ToFloat() (ScalarValue, error) {
	unix := time.Time(d).Unix()
	return Float(float64(unix)), nil
}

func (d Date) Equal(other Value) (bool, error) {
	x, ok := other.(Date)
	if !ok {
		return false, ErrCompatible
	}
	return time.Time(d).Equal(time.Time(x)), nil
}

func (d Date) Less(other Value) (bool, error) {
	x, ok := other.(Date)
	if !ok {
		return false, ErrCompatible
	}
	return time.Time(d).Before(time.Time(x)), nil
}

type Float float64

func (Float) Type() string {
	return "number"
}

func (Float) Kind() ValueKind {
	return KindScalar
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Scalar() any {
	return float64(f)
}

func (f Float) Add(other Value) (ScalarValue, error) {
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f + x), nil
}

func (f Float) Sub(other Value) (ScalarValue, error) {
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f - x), nil
}

func (f Float) Mul(other Value) (ScalarValue, error) {
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f * x), nil
}

func (f Float) Div(other Value) (ScalarValue, error) {
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	if x == 0 {
		return ErrDiv0, nil
	}
	return Float(f / x), nil
}

func (f Float) Pow(other Value) (ScalarValue, error) {
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(math.Pow(float64(f), float64(x))), nil
}

func (f Float) ToString() (ScalarValue, error) {
	return Text(f.String()), nil
}

func (f Float) ToBool() (ScalarValue, error) {
	return Boolean(float64(f) != 0), nil
}

func (f Float) ToFloat() (ScalarValue, error) {
	return f, nil
}

func (f Float) Equal(other Value) (bool, error) {
	x, ok := other.(Float)
	if !ok {
		return false, ErrCompatible
	}
	return float64(f) == float64(x), nil
}

func (f Float) Less(other Value) (bool, error) {
	x, ok := other.(Float)
	if !ok {
		return false, ErrCompatible
	}
	return float64(f) < float64(x), nil
}

type Text string

func (Text) Type() string {
	return "text"
}

func (Text) Kind() ValueKind {
	return KindScalar
}

func (t Text) String() string {
	return string(t)
}

func (t Text) Scalar() any {
	return string(t)
}

func (t Text) Add(other Value) (ScalarValue, error) {
	f, err := CastToFloat(t)
	if err != nil {
		return ErrValue, err
	}
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f + x), nil
}

func (t Text) Sub(other Value) (ScalarValue, error) {
	f, err := CastToFloat(t)
	if err != nil {
		return ErrValue, err
	}
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f - x), nil
}

func (t Text) Mul(other Value) (ScalarValue, error) {
	f, err := CastToFloat(t)
	if err != nil {
		return ErrValue, err
	}
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(f * x), nil
}

func (t Text) Div(other Value) (ScalarValue, error) {
	f, err := CastToFloat(t)
	if err != nil {
		return ErrValue, err
	}
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	if x == 0 {
		return ErrDiv0, nil
	}
	return Float(f * x), nil
}

func (t Text) Pow(other Value) (ScalarValue, error) {
	f, err := CastToFloat(t)
	if err != nil {
		return ErrValue, err
	}
	x, err := CastToFloat(other)
	if err != nil {
		return ErrValue, err
	}
	return Float(math.Pow(float64(f), float64(x))), nil
}

func (t Text) Concat(other Value) (ScalarValue, error) {
	x, err := CastToText(other)
	if err != nil {
		return nil, err
	}
	return Text(t + x), nil
}

func (t Text) ToString() (ScalarValue, error) {
	return t, nil
}

func (t Text) ToBool() (ScalarValue, error) {
	return Boolean(string(t) != ""), nil
}

func (t Text) ToFloat() (ScalarValue, error) {
	n, err := strconv.ParseFloat(string(t), 64)
	if err != nil {
		return ErrNA, nil
	}
	return Float(n), nil
}

func (t Text) Equal(other Value) (bool, error) {
	x, ok := other.(Text)
	if !ok {
		return false, ErrCompatible
	}
	return string(t) == string(x), nil
}

func (t Text) Less(other Value) (bool, error) {
	x, ok := other.(Text)
	if !ok {
		return false, ErrCompatible
	}
	return string(t) < string(x), nil
}

type Boolean bool

func (Boolean) Type() string {
	return "boolean"
}

func (Boolean) Kind() ValueKind {
	return KindScalar
}

func (b Boolean) String() string {
	return strconv.FormatBool(bool(b))
}

func (b Boolean) Scalar() any {
	return bool(b)
}

func (b Boolean) ToString() (ScalarValue, error) {
	s := strconv.FormatBool(bool(b))
	return Text(s), nil
}

func (b Boolean) ToBool() (ScalarValue, error) {
	return b, nil
}

func (b Boolean) ToFloat() (ScalarValue, error) {
	if !bool(b) {
		return Float(0), nil
	}
	return Float(1), nil
}

func (b Boolean) Equal(other Value) (bool, error) {
	x, ok := other.(Boolean)
	if !ok {
		return false, ErrCompatible
	}
	return bool(b) == bool(x), nil
}

func (b Boolean) Less(other Value) (bool, error) {
	x, ok := other.(Boolean)
	if !ok {
		return false, ErrCompatible
	}
	return bool(b) && !bool(x), nil
}
