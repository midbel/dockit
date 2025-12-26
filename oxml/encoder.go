package oxml

import (
	"io"

	"github.com/midbel/dockit/csv"
)

type Encoder interface {
	EncodeSheet(*Sheet) error
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

func (e *csvEncoder) EncodeSheet(sheet *Sheet) error {
	writer := csv.NewWriter(e.writer)
	writer.Comma = e.comma
	// for row := range sheet.Iter() {
	// 	if err := writer.Write(row); err != nil {
	// 		return err
	// 	}
	// }
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

func (e *jsonEncoder) EncodeSheet(sheet *Sheet) error {
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

func (e *xmlEncoder) EncodeSheet(sheet *Sheet) error {
	return nil
}
