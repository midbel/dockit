package gridx

import (
	"strings"
	"testing"

	"github.com/midbel/dockit/grid"
	"github.com/midbel/dockit/internal/testutil"
)

const firstSample = `go,foo,100
go,foo,100
go,bar,50
js,quz,100`
const secondSample = `go,foo,250
go,foo,100
js,bar,100`
const thirdSample = `be,1000
fr,5000
de,8000`

func TestUnion(t *testing.T) {
	t.Run("same-sample", testUnionSameSample)
	t.Run("different-sample", testUnionDifferentSample)
	t.Run("columns-mismatched", testUnionMismatchedColumns)
}

func testUnionMismatchedColumns(t *testing.T) {
	var (
		v1 = getViewFrom(t, firstSample)
		v2 = getViewFrom(t, thirdSample)
	)
	_, err := Union(v1, v2)
	if err == nil {
		t.Fatalf("error expected for mismatched columns count")
	}
}

func testUnionSameSample(t *testing.T) {
	var (
		v1 = getViewFrom(t, firstSample)
		v2 = getViewFrom(t, firstSample)
	)
	view, err := Union(v1, v2)
	if err != nil {
		t.Fatalf("error with union of two views: %s", err)
	}
	want := [][]string{
		{"go", "foo", "100"},
		{"go", "bar", "50"},
		{"js", "quz", "100"},
	}
	got := testutil.Collect(view)
	testutil.AssertSize(t, view, got)
	testutil.AssertViewEqual(t, want, got, nil)
}

func testUnionDifferentSample(t *testing.T) {
	var (
		v1 = getViewFrom(t, firstSample)
		v2 = getViewFrom(t, secondSample)
	)
	view, err := Union(v1, v2)
	if err != nil {
		t.Fatalf("error with union of two views: %s", err)
	}
	want := [][]string{
		{"go", "foo", "100"},
		{"go", "bar", "50"},
		{"js", "quz", "100"},
		{"go", "foo", "250"},
		{"js", "bar", "100"},
	}
	got := testutil.Collect(view)
	testutil.AssertSize(t, view, got)
	testutil.AssertViewEqual(t, want, got, nil)
}

func TestIntersect(t *testing.T) {
	var (
		v1 = getViewFrom(t, firstSample)
		v2 = getViewFrom(t, secondSample)
	)
	view, err := Intersect(v1, v2)
	if err != nil {
		t.Fatalf("error with intersect of two views: %s", err)
	}
	want := [][]string{
		{"go", "foo", "100"},
	}
	got := testutil.Collect(view)
	testutil.AssertSize(t, view, got)
	testutil.AssertViewEqual(t, want, got, nil)
}

func TestExcept(t *testing.T) {
	var (
		v1 = getViewFrom(t, firstSample)
		v2 = getViewFrom(t, secondSample)
	)
	view, err := Except(v1, v2)
	if err != nil {
		t.Fatalf("error with except of two views: %s", err)
	}
	want := [][]string{
		{"go", "bar", "50"},
		{"js", "quz", "100"},
	}
	got := testutil.Collect(view)
	testutil.AssertSize(t, view, got)
	testutil.AssertViewEqual(t, want, got, nil)
}

func getViewFrom(t *testing.T, content string) grid.View {
	t.Helper()

	f, err := testutil.CreateCsvFile(strings.NewReader(content))
	if err != nil {
		t.Fatalf("fail to create file from sample: %s", err)
	}
	sh, err := f.ActiveSheet()
	if err != nil {
		t.Fatalf("fail to retrieve active sheet: %s", err)
	}
	return sh
}
