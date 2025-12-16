package oxml

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

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
	for row := range sheet.Iter() {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

type jsonEncoder struct {
	writer  io.Writer
	AsArray bool
}

func EncodeJSON(w io.Writer) Encoder {
	return &jsonEncoder{
		writer: w,
	}
}

func (e *jsonEncoder) EncodeSheet(sheet *Sheet) error {
	var (
		data any
		err  error
	)
	if e.AsArray {
		data, err = e.getArray(sheet)
	} else {
		data, err = e.getObject(sheet)
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(e.writer).Encode(data)
}

func (e *jsonEncoder) getArray(sheet *Sheet) ([][]any, error) {
	var data [][]any
	for _, rs := range sheet.Rows {
		data = append(data, rs.values())
	}
	return data, nil
}

func (e *jsonEncoder) getObject(sheet *Sheet) ([]any, error) {
	if len(sheet.Rows) <= 1 {
		return nil, nil
	}
	var (
		ptr  = createType(sheet.Rows[0].Data(), "json")
		data []any
	)
	for i := 1; i < len(sheet.Rows); i++ {
		v := reflect.New(ptr).Elem()
		for i, str := range sheet.Rows[i].Data() {
			v.Field(i).SetString(str)
		}
		data = append(data, v.Addr().Interface())
	}
	return data, nil
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

func createType(names []string, format string) reflect.Type {
	var fields []reflect.StructField
	for _, n := range names {
		t := fmt.Sprintf("%s:\"%s\"", format, strings.ToLower(n))
		s := reflect.StructField{
			Name: strings.ToTitle(n),
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(t),
		}
		fields = append(fields, s)
	}
	return reflect.StructOf(fields)
}
