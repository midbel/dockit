package doc

import (
	"os"

	"github.com/midbel/dockit/csv"
	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/oxml"
)

func Infos(file string) ([]grid.ViewInfo, error) {
	f, err := Open(file)
	if err != nil {
		return nil, err
	}
	return f.Infos(), nil
}

func Open(file string) (grid.File, error) {
	format, err := detectFormat(file)
	if err != nil {
		return nil, err
	}
	switch format {
	case CSV:
		return csv.Open(file)
	case OXML:
		return oxml.Open(file)
	case ODS:
	default:
	}
	return nil, nil
}

type Format int

const (
	CSV Format = iota << 1
	OXML
	ODS
	Unknown
)

func detectFormat(file string) (Format, error) {
	r, err := os.Open(file)
	if err != nil {
		return Unknown, err
	}
	defer r.Close()
	return Unknown, nil
}
