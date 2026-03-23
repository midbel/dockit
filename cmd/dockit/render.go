package main

import (
	"io"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/csv"
)

type CsvRenderer struct {
	out       io.Writer
	Delimiter byte
	Quoted    bool
}

func NewCsvRenderer(w io.Writer) *CsvRenderer {
	return &CsvRenderer{
		out:       w,
		Delimiter: ',',
		Quoted:    false,
	}
}

func (r *CsvRenderer) Render(tbl cli.Table) error {
	ws := csv.NewWriter(r.out)
	ws.ForceQuote = r.Quoted
	ws.Comma = r.Delimiter

	defer ws.Flush()

	ws.Write(tbl.Headers)
	for _, r := range tbl.Rows {
		if err := ws.Write(r); err != nil {
			return err
		}
	}
	return nil
}
