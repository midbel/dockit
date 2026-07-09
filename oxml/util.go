package oxml

import (
	"github.com/midbel/dockit/value"
)

func typeFromValue(val value.Value) string {
	switch val.Type() {
	case value.TypeNumber:
		return TypeNumber
	case value.TypeText:
		return TypeInlineStr
	case value.TypeBool:
		return TypeBool
	case value.TypeDate:
		return TypeDate
	default:
		return TypeInlineStr
	}
}
