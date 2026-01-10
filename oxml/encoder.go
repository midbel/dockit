package oxml

import (
	"io"

	"github.com/midbel/dockit/csv"
)

type Encoder interface {
	EncodeSheet(View) error
}

type csvEncoder struct {
	writer io.Writer
	comma  byte
}

func EncodeCSV(w io.Writer) Encoder {
	return &csvEncoder{
		writer: w,
		comma:  ',',
	}
}

func (e *csvEncoder) EncodeSheet(it View) error {
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

func EncodeJSON(w io.Writer) Encoder {
	return &jsonEncoder{
		writer: w,
	}
}

func (e *jsonEncoder) EncodeSheet(it View) error {
	return nil
}

type xmlEncoder struct {
	writer io.Writer
}

func EncodeXML(w io.Writer) Encoder {
	return &xmlEncoder{
		writer: w,
	}
}

func (e *xmlEncoder) EncodeSheet(it View) error {
	return nil
}
