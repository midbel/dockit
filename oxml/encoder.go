package oxml

import (
	"io"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
)


type csvEncoder struct {
	writer io.Writer
	comma  byte
}

func EncodeCSV(w io.Writer) grid.Encoder {
	return &csvEncoder{
		writer: w,
		comma:  ',',
	}
}

func (e *csvEncoder) EncodeSheet(it grid.View) error {
	writer := csv.NewWriter(e.writer)
	writer.Comma = e.comma
	writer.ForceQuote = true
	for row := range it.Rows() {
		var fields []string
		for i := range row {
			fields = append(fields, row[i].String())
		}
		if err := writer.Write(fields); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

type jsonEncoder struct {
	writer io.Writer
}

func EncodeJSON(w io.Writer) grid.Encoder {
	return &jsonEncoder{
		writer: w,
	}
}

func (e *jsonEncoder) EncodeSheet(it grid.View) error {
	return nil
}

type xmlEncoder struct {
	writer io.Writer
}

func EncodeXML(w io.Writer) grid.Encoder {
	return &xmlEncoder{
		writer: w,
	}
}

func (e *xmlEncoder) EncodeSheet(it grid.View) error {
	return nil
}
