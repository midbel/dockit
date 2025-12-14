package oxml

import (
	"encoding/csv"
	"encoding/json"
	"io"
)

type Encoder interface {
	EncodeSheet(*Sheet) error
}

type csvEncoder struct {
	writer io.Writer
	comma  rune
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
	for row := range sheet.Iter() {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

type jsonEncoder struct {
	writer io.Writer
	fields []string
}

func EncodeJSON(w io.Writer) Encoder {
	return &jsonEncoder{
		writer: w,
	}
}

func (e *jsonEncoder) EncodeSheet(sheet *Sheet) error {
	var data [][]any
	for _, rs := range sheet.Rows {
		data = append(data, rs.values())
	}
	return json.NewEncoder(e.writer).Encode(data)
}

type xmlEncoder struct {
	writer io.Writer
	fields []string
}

func EncodeXML(w io.Writer) Encoder {
	return &xmlEncoder{
		writer: w,
	}
}

func (e *xmlEncoder) EncodeSheet(sheet *Sheet) error {
	return nil
}
