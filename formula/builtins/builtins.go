package builtins

import (
	"errors"

	"github.com/midbel/dockit/value"
)

var ErrArity = errors.New("invalid number of arguments")

var Registry = map[string]func([]value.Value) (value.Value, error){
	"min":    Min,
	"max":    Max,
	"sum":    Sum,
	"avg":    Avg,
	"count":  Count,
	"typeof": TypeOf,
}
