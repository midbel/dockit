package main

import (
	"io"

	"github.com/midbel/cli"
	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/grid/format"
	"github.com/midbel/dockit/value"
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

func sheet2Table(sheet grid.View, skipErr bool) cli.Table {
	var (
		t cli.Table
		i int
	)
	ft := createFormatter()
	for _, r := range sheet.Rows() {
		var (
			row     = make([]string, 0, len(r))
			discard bool
		)
		for _, v := range r {
			if value.IsError(v) && skipErr {
				discard = true
				break
			}
			str, _ := ft.Format(v)
			row = append(row, str)
		}
		if discard {
			continue
		}
		if i == 0 {
			t.Headers = row
		} else {
			t.Rows = append(t.Rows, row)
		}
		i++
	}
	return t
}

func createFormatter() format.Formatter {
	nft, err := format.ParseNumberFormatter("#######.##")
	if err != nil {
		nft = format.DefaultNumberFormatter()
	}
	ft := format.FormatValue()
	ft.Set(value.TypeNumber, nft)
	ft.Set(value.TypeBool, format.FormatBool())
	ft.Set(value.TypeText, format.FormatString())
	ft.Set(value.TypeDate, format.DefaultDateFormatter())

	return ft
}
