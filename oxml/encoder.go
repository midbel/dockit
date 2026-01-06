package oxml

import (
	"io"
	"iter"

	"github.com/midbel/dockit/csv"
)

type RowIterator interface {
	Rows() iter.Seq[[]any]
}

type Encoder interface {
	EncodeSheet(RowIterator) error
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

func (e *csvEncoder) EncodeSheet(it RowIterator) error {
	writer := csv.NewWriter(e.writer)
	writer.Comma = e.comma
	writer.ForceQuote = true
	for row := range it.Rows() {
		var fields []string
		for i := range row {
			fields = append(fields, valueToString(row[i]))
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

func (e *jsonEncoder) EncodeSheet(it RowIterator) error {
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

func (e *xmlEncoder) EncodeSheet(it RowIterator) error {
	return nil
}
