package flat

import (
	"errors"
	"io"
	"os"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/workbook"
)

type csvLoader struct {
	delimiter byte
	name      string
}

func NewCommaLoader() workbook.Loader {
	return createLoader(',', "comma")
}

func NewTabLoader() workbook.Loader {
	return createLoader('\t', "tab")
}

func NewSemicolonLoader() workbook.Loader {
	return createLoader(';', "semi")
}

func NewColonLoader() workbook.Loader {
	return createLoader(':', "colon")
}

func createLoader(delimiter byte, name string) workbook.Loader {
	return csvLoader{
		delimiter: delimiter,
		name:      name,
	}
}

func (x csvLoader) Name() string {
	return x.name
}

func (x csvLoader) Detect(file string) (bool, error) {
	r, err := os.Open(file)
	if err != nil {
		return false, nil
	}
	rs := csv.NewReader(r)
	rs.Comma = x.delimiter

	var (
		rows = make(map[int]int)
		iter = 16
	)
	for i := 0; i < iter; i++ {
		rec, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				iter = i
				break
			}
			return false, err
		}
		if len(rec) <= 1 {
			continue
		}
		rows[len(rec)]++
	}
	if len(rows) == 0 {
		return false, nil
	}
	var best int
	for _, f := range rows {
		best = max(best, f)
	}
	if len(rows) == 1 {
		return true, nil
	}
	return float64(len(rows))/float64(iter) >= 0.6 && best > 2, nil
}

func (x csvLoader) Open(file string) (grid.File, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	rs := csv.NewReader(r)
	rs.Comma = x.delimiter

	return OpenReader(rs)
}
