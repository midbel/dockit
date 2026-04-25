package testutil

import (
	"slices"
	"testing"
)

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
