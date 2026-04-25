package testutil

import (
	"slices"
	"testing"

	"github.com/midbel/dockit/grid"
)

func AssertSize(t *testing.T, view grid.View, got [][]string) {
	t.Helper()

	bd := view.Bounds()
	if bd.Height() != int64(len(got)) {
		t.Fatalf("number of rows mismatched! want %d, got %d", bd.Height(), len(got))
	}
	for i := range got {
		if int64(len(got[i])) != bd.Width() {
			t.Fatalf("row #%d: number of columns mismatched! want %d, got %d", i+1, bd.Width(), len(got[i]))
		}
	}
}

func AssertViewEqual(t *testing.T, want, got [][]string, cmp func([]string, []string) int) {
	t.Helper()

	if cmp != nil {
		slices.SortFunc(got, cmp)
		slices.SortFunc(want, cmp)
	}

	if len(got) != len(want) {
		t.Fatalf("number of rows mismatched! want %d, got %d", len(want), len(got))
		return
	}
	for i := 0; i < len(got); i++ {
		if !slices.Equal(want[i], got[i]) {
			t.Errorf("[%d] results mismatched! want %v, got %v", i, want[i], got[i])
		}
	}
}
