package builtins

import (
	"bufio"
	"io"
	"strings"

	"github.com/midbel/textwrap"
)

func Help(w io.Writer, ident string) error {
	b, err := Get(ident)
	if err != nil {
		return err
	}
	ws := bufio.NewWriter(w)
	defer ws.Flush()
	io.WriteString(ws, strings.ToUpper(b.Name))
	io.WriteString(ws, "(")
	for i, p := range b.Params {
		if i > 0 {
			io.WriteString(ws, ", ")
		}
		if p.Optional {
			io.WriteString(ws, "[")
		}
		io.WriteString(ws, strings.ToUpper(p.Name))
		io.WriteString(ws, ": ")
		io.WriteString(ws, p.Type)
		if p.Variadic {
			io.WriteString(ws, "...")
		}
		if p.Optional {
			io.WriteString(ws, "]")
		}
	}
	io.WriteString(ws, ")")

	io.WriteString(ws, "\t[")
	io.WriteString(ws, b.Category)
	io.WriteString(ws, "]")
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")

	io.WriteString(ws, textwrap.Wrap(b.Desc, 72))
	io.WriteString(ws, ".\n")
	io.WriteString(ws, "\n")

	io.WriteString(ws, "PARAMETERS:")
	io.WriteString(ws, "\n")
	for _, p := range b.Params {
		io.WriteString(ws, " "+strings.ToUpper(p.Name))
		io.WriteString(ws, ": ")
		io.WriteString(ws, p.Desc)
		io.WriteString(ws, "\n")
	}
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")
	io.WriteString(ws, "SUPPORTED BY: ")
	supports := map[string]bool{
		"ODS":  b.OdsSupported(),
		"OXML": b.OxmlSupported(),
	}
	var lino int
	for g, ok := range supports {
		if lino > 0 {
			io.WriteString(ws, ", ")
		}
		if ok {
			io.WriteString(ws, g)
			lino++
		}
	}
	io.WriteString(ws, "\n")
	io.WriteString(ws, "\n")
	return nil
}
