package diff

import (
	_ "embed"
	"strings"
	"testing"
)

type algorithmFactory struct {
	name string
	new  func() Diff
}

var diffAlgorithms = []algorithmFactory{
	{name: "LCS", new: func() Diff { return &LCS{} }},
	{name: "Myers", new: func() Diff { return &Myers{} }},
}

//go:embed fixtures/diffs_original.md
var fixtureOriginal string

//go:embed fixtures/diffs_updated.md
var fixtureUpdated string

func TestDiff_Compute_EmptySequences(t *testing.T) {
	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()

			t.Run("both empty", func(t *testing.T) {
				edits, err := m.Compute([]string{}, []string{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(edits) != 0 {
					t.Errorf("expected 0 edits, got %d", len(edits))
				}
			})

			t.Run("a empty, b has content", func(t *testing.T) {
				b := []string{"line1", "line2"}
				edits, err := m.Compute([]string{}, b)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(edits) != 2 {
					t.Fatalf("expected 2 edits, got %d", len(edits))
				}
				for i, edit := range edits {
					if edit.Kind != Insert {
						t.Errorf("edit %d: expected Insert, got %v", i, edit.Kind)
					}
					if edit.Content != b[i] {
						t.Errorf("edit %d: expected content %q, got %q", i, b[i], edit.Content)
					}
				}
			})

			t.Run("b empty, a has content", func(t *testing.T) {
				a := []string{"line1", "line2"}
				edits, err := m.Compute(a, []string{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(edits) != 2 {
					t.Fatalf("expected 2 edits, got %d", len(edits))
				}
				for i, edit := range edits {
					if edit.Kind != Delete {
						t.Errorf("edit %d: expected Delete, got %v", i, edit.Kind)
					}
					if edit.Content != a[i] {
						t.Errorf("edit %d: expected content %q, got %q", i, a[i], edit.Content)
					}
				}
			})
		})
	}
}

func TestDiff_Compute_IdenticalSequences(t *testing.T) {
	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "line2", "line3"}

	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(edits) != 3 {
				t.Fatalf("expected 3 edits, got %d", len(edits))
			}

			for i, edit := range edits {
				if edit.Kind != Equal {
					t.Errorf("edit %d: expected Equal, got %v", i, edit.Kind)
				}
				if edit.AIndex != i || edit.BIndex != i {
					t.Errorf("edit %d: expected indices (%d,%d), got (%d,%d)", i, i, i, edit.AIndex, edit.BIndex)
				}
			}
		})
	}
}

func TestDiff_Compute_SimpleInsert(t *testing.T) {
	a := []string{"line1", "line3"}
	b := []string{"line1", "line2", "line3"}

	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify structure: Equal(line1), Insert(line2), Equal(line3)
			if len(edits) != 3 {
				t.Fatalf("expected 3 edits, got %d", len(edits))
			}

			if edits[0].Kind != Equal || edits[0].Content != "line1" {
				t.Errorf("edit 0: expected Equal(line1), got %v(%s)", edits[0].Kind, edits[0].Content)
			}
			if edits[1].Kind != Insert || edits[1].Content != "line2" {
				t.Errorf("edit 1: expected Insert(line2), got %v(%s)", edits[1].Kind, edits[1].Content)
			}
			if edits[2].Kind != Equal || edits[2].Content != "line3" {
				t.Errorf("edit 2: expected Equal(line3), got %v(%s)", edits[2].Kind, edits[2].Content)
			}
		})
	}
}

func TestDiff_Compute_SimpleDelete(t *testing.T) {
	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "line3"}

	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify structure: Equal(line1), Delete(line2), Equal(line3)
			if len(edits) != 3 {
				t.Fatalf("expected 3 edits, got %d", len(edits))
			}

			if edits[0].Kind != Equal || edits[0].Content != "line1" {
				t.Errorf("edit 0: expected Equal(line1), got %v(%s)", edits[0].Kind, edits[0].Content)
			}
			if edits[1].Kind != Delete || edits[1].Content != "line2" {
				t.Errorf("edit 1: expected Delete(line2), got %v(%s)", edits[1].Kind, edits[1].Content)
			}
			if edits[2].Kind != Equal || edits[2].Content != "line3" {
				t.Errorf("edit 2: expected Equal(line3), got %v(%s)", edits[2].Kind, edits[2].Content)
			}
		})
	}
}

func TestDiff_Compute_CompleteReplacement(t *testing.T) {
	a := []string{"old1", "old2"}
	b := []string{"new1", "new2"}

	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should be all deletes followed by all inserts (or interleaved)
			deleteCount := 0
			insertCount := 0
			for _, edit := range edits {
				switch edit.Kind {
				case Delete:
					deleteCount++
				case Insert:
					insertCount++
				case Equal:
					t.Errorf("unexpected Equal edit when sequences are completely different")
				}
			}

			if deleteCount != 2 {
				t.Errorf("expected 2 deletes, got %d", deleteCount)
			}
			if insertCount != 2 {
				t.Errorf("expected 2 inserts, got %d", insertCount)
			}
		})
	}
}

func TestDiff_Compute_Fixtures(t *testing.T) {
	original := strings.Split(strings.TrimSpace(fixtureOriginal), "\n")
	updated := strings.Split(strings.TrimSpace(fixtureUpdated), "\n")

	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(original, updated)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(edits) == 0 {
				t.Fatal("expected non-empty edit list")
			}

			reconstructed := ApplyEdits(original, edits)
			if len(reconstructed) != len(updated) {
				t.Fatalf("reconstructed length %d != updated length %d", len(reconstructed), len(updated))
			}
			for i := range reconstructed {
				if reconstructed[i] != updated[i] {
					t.Errorf("line %d: reconstructed %q != updated %q", i, reconstructed[i], updated[i])
				}
			}

			counts := CountEditKinds(edits)
			if counts[Equal] == 0 {
				t.Error("expected some Equal edits (files share common lines like blank lines)")
			}
			if counts[Insert] == 0 {
				t.Error("expected some Insert edits")
			}
			if counts[Delete] == 0 {
				t.Error("expected some Delete edits")
			}

			t.Logf("Edit statistics: Equal=%d, Insert=%d, Delete=%d, Total=%d",
				counts[Equal], counts[Insert], counts[Delete], len(edits))
		})
	}
}

func TestDiff_Name(t *testing.T) {
	for _, alg := range diffAlgorithms {
		alg := alg
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			if m.Name() != alg.name {
				t.Errorf("expected name %q, got %q", alg.name, m.Name())
			}
		})
	}
}
