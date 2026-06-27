package builtins

import (
	"testing"

	"github.com/midbel/dockit/value"
)

func TestStrings(t *testing.T) {
	t.Run("len", testLen)
	t.Run("upper", testUpper)
	t.Run("lower", testLower)
}

func testLen(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Float(6),
		},
		{
			Args: []value.Value{value.Text("")},
			Want: value.Float(0),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, Len, tests)
}

func testUpper(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Text("FOOBAR"),
		},
		{
			Args: []value.Value{value.Text("FOOBAR")},
			Want: value.Text("FOOBAR"),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, Upper, tests)
}

func testLower(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.Text("FOOBAR")},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, Lower, tests)
}