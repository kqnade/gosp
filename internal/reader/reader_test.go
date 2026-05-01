package reader

import (
	"testing"

	"github.com/kqnade/gosp/internal/value"
)

func TestReadAllReadsMultipleForms(t *testing.T) {
	got, err := ReadAll("a (b c)")
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	want := []value.Value{
		value.Symbol{Name: "a"},
		value.List(value.Symbol{Name: "b"}, value.Symbol{Name: "c"}),
	}
	if len(got) != len(want) {
		t.Fatalf("ReadAll returned %d forms, want %d", len(got), len(want))
	}
	for i := range want {
		if !valuesEqual(got[i], want[i]) {
			t.Fatalf("ReadAll[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestReadAllSkipsCommentsAndWhitespace(t *testing.T) {
	got, err := ReadAll(" ; comment\n a \n\n b ")
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	want := []value.Value{
		value.Symbol{Name: "a"},
		value.Symbol{Name: "b"},
	}
	if len(got) != len(want) {
		t.Fatalf("ReadAll returned %d forms, want %d", len(got), len(want))
	}
	for i := range want {
		if !valuesEqual(got[i], want[i]) {
			t.Fatalf("ReadAll[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
