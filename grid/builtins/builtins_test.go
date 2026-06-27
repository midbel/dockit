package builtins

import (
	"testing"

	"github.com/midbel/dockit/value"
)

type BuiltinTestCase struct {
	Args []value.Value
	Want value.Value
}

func testBuiltin(t *testing.T, fn BuiltinFunc, args []BuiltinTestCase) {
	t.Helper()
	for _, tt := range args {
		got := fn(tt.Args)
		ok := value.Eq(tt.Want, got)
		if !value.True(ok) {
			t.Errorf("result mismatched! want %v, got %v", tt.Want, got)
		}
	}
}
