package builtins

import (
	"testing"

	"github.com/midbel/dockit/value"
)

func TestStrings(t *testing.T) {
	t.Run("len", testLen)
	t.Run("upper", testUpper)
	t.Run("lower", testLower)
	t.Run("concat", testConcat)
	t.Run("left", testLeft)
	t.Run("right", testRight)
	t.Run("mid", testMid)
	t.Run("trim", testTrim)
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

func testConcat(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foo"), value.Text("bar")},
			Want: value.Text("foobar"),
		},
	}
	testBuiltin(t, Concat, tests)
}

func testLeft(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foo"), value.Float(1)},
			Want: value.Text("f"),
		},
		{
			Args: []value.Value{value.Text("foo")},
			Want: value.Text("f"),
		},
		{
			Args: []value.Value{value.Text("foo"), value.Float(3)},
			Want: value.Text("foo"),
		},
		{
			Args: []value.Value{value.Text("foo"), value.Float(7)},
			Want: value.Text("foo"),
		},
	}
	testBuiltin(t, Left, tests)
}

func testRight(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foo"), value.Float(1)},
			Want: value.Text("o"),
		},
		{
			Args: []value.Value{value.Text("foo")},
			Want: value.Text("o"),
		},
		{
			Args: []value.Value{value.Text("foo"), value.Float(2)},
			Want: value.Text("oo"),
		},
		{
			Args: []value.Value{value.Text("foo"), value.Float(7)},
			Want: value.Text("foo"),
		},
	}
	testBuiltin(t, Right, tests)
}

func testMid(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foobar"), value.Float(1), value.Float(6)},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.Text("foobar"), value.Float(1), value.Float(3)},
			Want: value.Text("foo"),
		},
		{
			Args: []value.Value{value.Text("foobar"), value.Float(4), value.Float(3)},
			Want: value.Text("bar"),
		},
		{
			Args: []value.Value{value.Text("foobar"), value.Float(3), value.Float(1)},
			Want: value.Text("o"),
		},
	}
	testBuiltin(t, Mid, tests)
}

func testTrim(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.Text(" foobar")},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.Text("foobar ")},
			Want: value.Text("foobar"),
		},
		{
			Args: []value.Value{value.Text(" foo  bar ")},
			Want: value.Text("foo bar"),
		},
		{
			Args: []value.Value{value.Text("f oo  ba r")},
			Want: value.Text("f oo ba r"),
		},
	}
	testBuiltin(t, Trim, tests)
}
