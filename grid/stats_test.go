package grid_test

import (
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
)

func TestStat(t *testing.T) {
	file := testutil.CreateFile()
	tests := []struct {
		Name      string
		Cells     int
		Formulas  int
		Errors    int
		Constants int
	}{
		{
			Name:      "sheet1",
			Cells:     6,
			Formulas:  2,
			Constants: 4,
		},
		{
			Name:      "sheet2",
			Cells:     8,
			Formulas:  4,
			Constants: 4,
		},
		{
			Name:      "sheet3",
			Cells:     8,
			Formulas:  4,
			Constants: 4,
		},
	}
	for _, c := range tests {
		s, err := getStats(file, c.Name)
		if err != nil {
			t.Errorf("error retrieving sheet %s", c.Name)
			continue
		}
		if s.Cells != c.Cells {
			t.Errorf("%s: cells count mismatched! want %d, got %d", c.Name, c.Cells, s.Cells)
		}
		if s.Formulas != c.Formulas {
			t.Errorf("%s: formulas count mismatched! want %d, got %d", c.Name, c.Formulas, s.Formulas)
		}
		if s.Errors != c.Errors {
			t.Errorf("%s: errors count mismatched! want %d, got %d", c.Name, c.Errors, s.Errors)
		}
		if s.Constants != c.Constants {
			t.Errorf("%s: constants count mismatched! want %d, got %d", c.Name, c.Constants, s.Constants)
		}
	}
}

func getStats(f grid.File, name string) (grid.ViewStats, error) {
	var stat grid.ViewStats
	sh, err := f.Sheet(name)
	if err == nil {
		stat = grid.AnalyzeView(sh)
	}
	return stat, nil
}
