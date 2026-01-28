package eval

import (
	"github.com/midbel/dockit/layout"
)

type LValue interface {
	Set(value.Value) error
}