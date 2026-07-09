package value

import (
	"errors"
	"testing"

	"github.com/midbel/dockit/layout"
)

func TestScalarValues(t *testing.T) {
	tests := []struct {
		Name   string
		Value  ScalarValue
		Type   string
		String string
		Scalar any
	}{
		{
			Name:   "blank",
			Value:  Empty(),
			Type:   TypeBlank,
			String: "",
			Scalar: nil,
		},
		{
			Name:   "number",
			Value:  Float(42.5),
			Type:   TypeNumber,
			String: "42.5",
			Scalar: float64(42.5),
		},
		{
			Name:   "text",
			Value:  Text("dockit"),
			Type:   TypeText,
			String: "dockit",
			Scalar: "dockit",
		},
		{
			Name:   "boolean",
			Value:  Boolean(true),
			Type:   TypeBool,
			String: "true",
			Scalar: true,
		},
		{
			Name:   "error",
			Value:  ErrRef,
			Type:   TypeError,
			String: "#REF!",
			Scalar: "#REF!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Value.Kind() != KindScalar && tt.Value.Kind() != KindError {
				t.Fatalf("unexpected kind %d", tt.Value.Kind())
			}
			if got := tt.Value.Type(); got != tt.Type {
				t.Fatalf("type mismatch: want %s, got %s", tt.Type, got)
			}
			if got := tt.Value.String(); got != tt.String {
				t.Fatalf("string mismatch: want %q, got %q", tt.String, got)
			}
			if got := tt.Value.Scalar(); got != tt.Scalar {
				t.Fatalf("scalar mismatch: want %#v, got %#v", tt.Scalar, got)
			}
		})
	}
}

func TestCastsAndTruth(t *testing.T) {
	if got, err := CastToFloat(Text("12.5")); err != nil || got != Float(12.5) {
		t.Fatalf("text to float mismatch: want 12.5, got %s (%v)", got, err)
	}
	if _, err := CastToFloat(Text("nope")); !errors.Is(err, ErrCast) {
		t.Fatalf("expected cast error, got %v", err)
	}
	if got, err := CastToText(Text("dockit")); err != nil || got != Text("dockit") {
		t.Fatalf("text cast mismatch: want dockit, got %s (%v)", got, err)
	}
	if _, err := CastToText(Float(3)); !errors.Is(err, ErrCast) {
		t.Fatalf("expected float to text cast error, got %v", err)
	}
	if !True(Boolean(true)) {
		t.Fatalf("true boolean should be truthy")
	}
	if True(Boolean(false)) {
		t.Fatalf("false boolean should not be truthy")
	}
	if !True(Text("x")) {
		t.Fatalf("non-empty text should be truthy")
	}
	if True(Empty()) {
		t.Fatalf("blank should not be truthy")
	}
}

func TestArithmeticAndComparison(t *testing.T) {
	tests := []struct {
		Name  string
		Got   Value
		Want  Value
		Check func(Value, Value) bool
	}{
		{Name: "add", Got: Add(Float(2), Float(3)), Want: Float(5), Check: sameValue},
		{Name: "subtract", Got: Sub(Float(7), Float(4)), Want: Float(3), Check: sameValue},
		{Name: "multiply", Got: Mul(Float(6), Float(7)), Want: Float(42), Check: sameValue},
		{Name: "divide", Got: Div(Float(8), Float(2)), Want: Float(4), Check: sameValue},
		{Name: "divide by zero", Got: Div(Float(8), Float(0)), Want: ErrDiv0, Check: sameValue},
		{Name: "power", Got: Pow(Float(2), Float(3)), Want: Float(8), Check: sameValue},
		{Name: "concat", Got: Concat(Text("dock"), Text("it")), Want: Text("dockit"), Check: sameValue},
		{Name: "equal", Got: Eq(Float(2), Float(2)), Want: Boolean(true), Check: sameValue},
		{Name: "not equal", Got: Ne(Text("a"), Text("b")), Want: Boolean(true), Check: sameValue},
		{Name: "less", Got: Lt(Float(1), Float(2)), Want: Boolean(true), Check: sameValue},
		{Name: "greater", Got: Gt(Float(3), Float(2)), Want: Boolean(true), Check: sameValue},
		{Name: "incompatible add", Got: Add(Boolean(true), Float(1)), Want: ErrValue, Check: sameValue},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if !tt.Check(tt.Got, tt.Want) {
				t.Fatalf("value mismatch: want %s, got %s", tt.Want, tt.Got)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	if got := HasErrors(Float(1), nil, ErrName, ErrRef); got != ErrName {
		t.Fatalf("expected first error %s, got %s", ErrName, got)
	}
	if got := HasErrors(Float(1), Text("ok")); got != nil {
		t.Fatalf("expected no error, got %s", got)
	}
	err := NewErrorFromCode("#CUSTOM!")
	if !IsError(err) {
		t.Fatalf("custom error should be an error value")
	}
	if got := err.String(); got != "#CUSTOM!" {
		t.Fatalf("custom error mismatch: got %s", got)
	}
}

func TestArray(t *testing.T) {
	arr := NewArray(Rows(
		[]Value{Float(1), Text("a")},
		[]Value{Float(2), Text("b")},
	)).(Array)

	if got, want := arr.Dimension(), (layout.Dimension{Lines: 2, Columns: 2}); !got.Equal(want) {
		t.Fatalf("dimension mismatch: want %#v, got %#v", want, got)
	}
	if got := arr.Count(); got != 4 {
		t.Fatalf("count mismatch: want 4, got %d", got)
	}
	if got := arr.At(1, 1); !sameValue(got, Text("b")) {
		t.Fatalf("cell mismatch: want b, got %s", got)
	}
	if got := arr.At(9, 9); !IsBlank(got) {
		t.Fatalf("out of bounds should return blank, got %s", got)
	}

	clone := arr.Clone()
	clone.SetAt(0, 0, Float(99))
	if sameValue(arr.At(0, 0), clone.At(0, 0)) {
		t.Fatalf("clone mutation should not change original")
	}
	if arr.Equal(clone) {
		t.Fatalf("changed clone should not be equal to original")
	}
}

func TestArrayHelpers(t *testing.T) {
	filled := ScalarToArray(Text("x"), 2, 3).(Array)
	if got, want := filled.Dimension(), (layout.Dimension{Lines: 2, Columns: 3}); !got.Equal(want) {
		t.Fatalf("dimension mismatch: want %#v, got %#v", want, got)
	}
	for v := range filled.Values() {
		if !sameValue(v, Text("x")) {
			t.Fatalf("filled array mismatch: got %s", v)
		}
	}

	source := NewArray(Rows(
		[]Value{Float(1), Float(2)},
		[]Value{Float(3), Float(4)},
	))
	mapped, err := ApplyArrayWithScalar(source, Float(10), func(left, right Value) (Value, error) {
		return Add(left, right), nil
	})
	if err != nil {
		t.Fatalf("apply array with scalar failed: %v", err)
	}
	want := NewArray(Rows(
		[]Value{Float(11), Float(12)},
		[]Value{Float(13), Float(14)},
	)).(Array)
	if !mapped.(Array).Equal(want) {
		t.Fatalf("mapped array mismatch: want %#v, got %#v", want, mapped)
	}
}

func TestIterHelpers(t *testing.T) {
	arr := NewArray(Rows(
		[]Value{Float(1), Text("skip")},
		[]Value{Float(2), Float(3)},
	))
	args := []Value{Float(4), arr}

	got := Collect(args, func(v Value) (Float, bool) {
		f, ok := v.(Float)
		return f, ok
	})
	want := []Float{4, 1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("collected length mismatch: want %d, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("collected[%d] mismatch: want %s, got %s", i, want[i], got[i])
		}
	}

	if got := FindIndex([]Value{Text("a"), Text("b")}, Text("b")); !sameValue(got, Float(2)) {
		t.Fatalf("find index mismatch: want 2, got %s", got)
	}
	if got := Find([]Value{Text("a"), Text("b")}, Text("x")); got != ErrNA {
		t.Fatalf("find should return #N/A, got %s", got)
	}
}

func sameValue(got, want Value) bool {
	if got == nil || want == nil {
		return got == want
	}
	if got.Type() != want.Type() || got.Kind() != want.Kind() {
		return false
	}
	return got.String() == want.String()
}
