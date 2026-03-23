package builtins

import (
	"github.com/midbel/dockit/value"
)

func If(args []value.Value) value.Value {
	if err := value.HasErrors(args[:3]...); err != nil {
		return err
	}
	if value.True(args[0]) {
		return args[1]
	}
	return args[2]
}

func IfError(args []value.Value) value.Value {
	if value.IsError(args[0]) {
		return args[1]
	}
	return args[0]
}

func IfNA(args []value.Value) value.Value {
	if value.IsError(args[0]) && args[0] == value.ErrNA {
		return args[1]
	}
	return args[0]
}

func And(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		ok1 = value.True(args[0])
		ok2 = value.True(args[1])
	)
	return value.Boolean(ok1 && ok2)
}

func Or(args []value.Value) value.Value {
	if err := value.HasErrors(args[:2]...); err != nil {
		return err
	}
	var (
		ok1 = value.True(args[0])
		ok2 = value.True(args[1])
	)
	return value.Boolean(ok1 || ok2)
}

func Xor(args []value.Value) value.Value {
	return nil
}

func Not(args []value.Value) value.Value {
	if err := value.HasErrors(args[0]); err != nil {
		return err
	}
	ok := value.True(args[0])
	return value.Boolean(!ok)
}
