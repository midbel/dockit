package types

import (
	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/value"
)

type Module struct {
	name    string
	alias   string
	environ *env.Environment
}

func NewModule(name, alias string) value.Value {
	m := Module{
		name:    name,
		alias:   alias,
		environ: env.Empty(),
	}
	return &m
}

func (*Module) Type() string {
	return "module"
}

func (*Module) Kind() value.ValueKind {
	return value.KindObject
}

func (m *Module) String() string {
	return m.name
}

func (m *Module) Get(ident string) value.Value {
	return m.environ.Resolve(ident)
}
