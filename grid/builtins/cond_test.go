package builtins

import (
	"testing"

	"github.com/midbel/dockit/value"
)

func TestCond(t *testing.T) {
	t.Run("ifs", testIfs)
	t.Run("if", testIf)
	t.Run("ifError", testIfError)
	t.Run("ifNa", testIfNa)
	t.Run("and", testAnd)
	t.Run("or", testOr)
	t.Run("xor", testXor)
	t.Run("not", testNot)
	t.Run("switch", testSwitch)
	t.Run("choose", testChoose)
}

func testIfs(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Float(42),
				value.Boolean(false),
				value.Float(0),
			},
			Want: value.Float(42),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Float(0),
				value.Boolean(true),
				value.Float(42),
			},
			Want: value.Float(42),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Float(42),
				value.Boolean(false),
				value.ErrValue,
			},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, Ifs, tests)
}

func testIf(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Float(42),
				value.Float(0),
			},
			Want: value.Float(42),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Float(42),
				value.Float(0),
			},
			Want: value.Float(0),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
				value.ErrValue,
				value.Float(42),
			},
			Want: value.ErrValue,
		},
		{
			Args: []value.Value{
				value.ErrValue,
				value.Boolean(true),
				value.Float(42),
			},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, If, tests)
}

func testIfError(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(42),
				value.Float(0),
			},
			Want: value.Float(42),
		},
		{
			Args: []value.Value{
				value.ErrDiv0,
				value.Float(0),
			},
			Want: value.Float(0),
		},
	}
	testBuiltin(t, IfError, tests)
}

func testIfNa(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(42),
				value.Float(0),
			},
			Want: value.Float(42),
		},
		{
			Args: []value.Value{
				value.ErrDiv0,
				value.Float(0),
			},
			Want: value.ErrDiv0,
		},
		{
			Args: []value.Value{
				value.ErrNA,
				value.Float(0),
			},
			Want: value.Float(0),
		},
	}
	testBuiltin(t, IfNA, tests)
}

func testAnd(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(42),
				value.Float(0),
			},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Boolean(true),
			},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Boolean(true),
			},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, And, tests)
}

func testOr(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(42),
				value.Float(0),
			},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Boolean(true),
			},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Boolean(true),
			},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Boolean(false),
			},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, Or, tests)
}

func testXor(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
			},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{
				value.Boolean(true),
				value.Boolean(true),
			},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
				value.Boolean(false),
			},
			Want: value.Boolean(false),
		},
	}
	testBuiltin(t, Xor, tests)
}

func testNot(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Boolean(true),
			},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{
				value.Boolean(false),
			},
			Want: value.Boolean(true),
		},
	}
	testBuiltin(t, Not, tests)
}

func testChoose(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(1),
				value.Text("foo"),
				value.Text("bar"),
			},
			Want: value.Text("foo"),
		},
		{
			Args: []value.Value{
				value.Float(2),
				value.Text("foo"),
				value.Text("bar"),
			},
			Want: value.Text("bar"),
		},
		{
			Args: []value.Value{
				value.Float(0),
				value.Text("foo"),
				value.Text("bar"),
			},
			Want: value.ErrNA,
		},
	}
	testBuiltin(t, Choose, tests)
}

func testSwitch(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{
				value.Float(42),			
				value.Float(0),
				value.Text("zero"),
				value.Float(42),
				value.Text("answer"),
				value.Text("oups!"),
			},
			Want: value.Text("answer"),
		},
		{
			Args: []value.Value{
				value.Float(-1),			
				value.Float(0),
				value.Text("zero"),
				value.Float(42),
				value.Text("answer"),
				value.Text("oups!"),
			},
			Want: value.Text("oups!"),
		},
	}
	testBuiltin(t, Switch, tests)
}
