package builtins

import (
	"math"

	"github.com/midbel/dockit/value"
)

func Choose(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	f := math.Floor(asFloat(args[0])) - 1
	if int(f) < 0 || int(f) >= len(args)-1 {
		return value.ErrNA
	}
	return args[int(f)]
}

func Switch(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	var (
		base     = args[0]
		rest     = args[1:]
		fallback value.Value
	)
	if z := len(rest); z%2 == 1 {
		fallback = rest[z-1]
		rest = rest[:z]
	}
	for i := 0; i < len(rest); i += 2 {
		ok := value.Eq(base, rest[i])
		if value.True(ok) {
			return rest[i+1]
		}
	}
	if fallback != nil {
		return fallback
	}
	return value.ErrNA
}

func Match(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	return nil
}

func Index(args []value.Value) value.Value {
	if err := value.HasErrors(args...); err != nil {
		return err
	}
	return nil
}

func VLookup(args []value.Value) value.Value {
	return nil
}

func HLookup(args []value.Value) value.Value {
	return nil
}

func XLookup(args []value.Value) value.Value {
	return nil
}
