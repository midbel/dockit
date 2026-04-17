package gridx

import (
	"fmt"
	"strings"

	"github.com/midbel/dockit/value"
)

func keyFromRow(row []value.ScalarValue, cols []int64) string {
	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteRune('|')
		}
		k := createKey(row[c])
		b.WriteString(k)
	}
	return b.String()
}

func createKey(v value.Value) string {
	var prefix string
	switch v.Type() {
	case value.TypeNumber:
		prefix = "n"
	case value.TypeText:
		prefix = "s"
	case value.TypeBool:
		prefix = "b"
	case value.TypeDate:
		prefix = "d"
	case value.TypeError:
		prefix = "e"
	case value.TypeBlank:
		prefix = "b"
	default:
		prefix = "?"
	}
	return fmt.Sprintf("%s:%s", prefix, v.String())
}
