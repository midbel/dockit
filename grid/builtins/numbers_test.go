package builtins

import (
	"testing"

	"github.com/midbel/dockit/value"
)

func TestNumbers(t *testing.T) {
	t.Run("isOdd", testIsOdd)
	t.Run("isEven", testIsEven)
}

func testIsEven(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Float(1)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Float(2)},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, IsEven, tests)
}

func testIsOdd(t *testing.T) {
	tests := []BuiltinTestCase{
		{
			Args: []value.Value{value.Float(0)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.Float(1)},
			Want: value.Boolean(true),
		},
		{
			Args: []value.Value{value.Float(2)},
			Want: value.Boolean(false),
		},
		{
			Args: []value.Value{value.ErrValue},
			Want: value.ErrValue,
		},
	}
	testBuiltin(t, IsOdd, tests)
}
