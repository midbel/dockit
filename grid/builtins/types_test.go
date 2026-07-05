package builtins

import (
	"testing"
	"time"

	"github.com/midbel/dockit/value"
)

func TestTypes(t *testing.T) {
	t.Run("typeof", testType)
	t.Run("isnumber", testIsNumber)
	t.Run("istext", testIsText)
	t.Run("isblank", testIsBlank)
	t.Run("iserror", testIsError)
	t.Run("isna", testIsNa)
	t.Run("na", testNa)
	t.Run("err", testErr)
}

func testType(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Text("number"),
		},
		{
			Args: []value.Value{value.Text("text")},
			Want: value.Text("text"),
		},
		{
			Args: []value.Value{value.Boolean(true)},
			Want: value.Text("boolean"),
		},
		{
			Args: []value.Value{value.Empty()},
			Want: value.Text("blank"),
		},
		{
			Args: []value.Value{value.Date(time.Now())},
			Want: value.Text("date"),
		},
	}
	testBuiltin(t, TypeOf, tests)
}

func testIsNumber(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Float(42)},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Text("")},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Boolean(true)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Empty()},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, IsNumber, tests)
}

func testIsText(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Text("")},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Boolean(true)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Empty()},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, IsText, tests)
}

func testIsBlank(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Text("")},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Boolean(true)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Empty()},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, IsBlank, tests)
}

func testIsError(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Text("")},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Text("foobar")},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Boolean(true)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Empty()},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.ErrDiv0},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.ErrNA},
			Want: value.Boolean(true),
		},
	}
	testBuiltin(t, IsError, tests)
}

func testIsNa(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.ErrValue},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrDiv0},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrNA},
			Want: value.Boolean(true),
		},
	}
	testBuiltin(t, IsNA, tests)
}

func testNa(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Want: value.ErrNA,
		},
	}
	testBuiltin(t, Na, tests)
}

func testErr(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Text("Null")},
			Want: value.ErrNull,
		},
		{
			Args: []value.Value{value.Text("Div0")},
			Want: value.ErrDiv0,
		},
		{
			Args: []value.Value{value.Text("Value")},
			Want: value.ErrValue,
		},
		{
			Args: []value.Value{value.Text("Ref")},
			Want: value.ErrRef,
		},
		{
			Args: []value.Value{value.Text("Name")},
			Want: value.ErrName,
		},
		{
			Args: []value.Value{value.Text("Num")},
			Want: value.ErrNum,
		},
		{
			Args: []value.Value{value.Text("Other")},
			Want: value.ErrNA,
		},
	}
	testBuiltin(t, Err, tests)
}
