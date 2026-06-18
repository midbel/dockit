package types

import (
	"os"
	"strconv"
	"strings"

	"github.com/midbel/dockit/value"
)

type flagValue struct {
	name string
	args []string
}

func NewFlagValue(name string, args []string) value.Value {
	return flagValue{
		name: name,
		args: args,
	}
}

func (flagValue) Immutable() bool {
	return true
}

func (flagValue) Type() string {
	return "flag"
}

func (flagValue) Kind() value.ValueKind {
	return value.KindObject
}

func (v flagValue) String() string {
	return v.Type()
}

func (v flagValue) Get(name string) value.Value {
	switch name {
	case "arg0":
		return value.Text(v.name)
	case "argN":
		return value.Float(len(v.args))
	default:
		name = strings.TrimPrefix(name, "arg")
		ix, err := strconv.Atoi(name)
		if err != nil {
			return value.ErrValue
		}
		ix--
		if ix < 0 || ix >= len(v.args) {
			return value.ErrName
		}
		return value.Text(v.args[ix])
	}
}

type envValue struct{}

func NewEnvValue() value.Value {
	return envValue{}
}

func (envValue) Immutable() bool {
	return true
}

func (envValue) Type() string {
	return "env"
}

func (envValue) Kind() value.ValueKind {
	return value.KindObject
}

func (v envValue) String() string {
	return v.Type()
}

func (v envValue) Get(name string) value.Value {
	str := os.Getenv(strings.ToUpper(name))
	return value.Text(str)
}
