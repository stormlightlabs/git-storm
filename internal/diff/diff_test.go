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

func TestAreSimilarLines(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical lines",
			a:        "github.com/foo/bar v1.0.0",
			b:        "github.com/foo/bar v1.0.0",
			expected: true,
		},
		{
			name:     "similar package different version",
			a:        "github.com/charmbracelet/x/ansi v0.10.1 // indirect",
			b:        "github.com/charmbracelet/x/ansi v0.10.3 // indirect",
			expected: true,
		},
		{
			name:     "different packages",
			a:        "github.com/charmbracelet/x/term v0.2.1 // indirect",
			b:        "github.com/charmbracelet/x/exp/teatest v0.0.0-20251",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty",
			a:        "some content",
			b:        "",
			expected: false,
		},
		{
			name:     "completely different",
			a:        "package main",
			b:        "import fmt",
			expected: false,
		},
		{
			name:     "short common prefix",
			a:        "github.com/foo/bar",
			b:        "github.com/baz/qux",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := areSimilarLines(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("areSimilarLines(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMergeReplacements(t *testing.T) {
	tests := []struct {
		name     string
		input    []Edit
		expected []Edit
	}{
		{
			name:     "empty edits",
			input:    []Edit{},
			expected: []Edit{},
		},
		{
			name: "go.mod scenario - deletes followed by inserts with gap",
			input: []Edit{
				{Kind: Delete, AIndex: 17, BIndex: -1, Content: "    github.com/charmbracelet/colorprofile v0.3.2 // indirect"},
				{Kind: Delete, AIndex: 18, BIndex: -1, Content: "    github.com/charmbracelet/lipgloss/v2"},
				{Kind: Delete, AIndex: 19, BIndex: -1, Content: "    github.com/charmbracelet/ultraviolet"},
				{Kind: Delete, AIndex: 20, BIndex: -1, Content: "    github.com/charmbracelet/x/ansi v0.10.1 // indirect"},
				{Kind: Insert, AIndex: -1, BIndex: 23, Content: "    github.com/aymanbagabas/go-udiff v0.3.1 // indirect"},
				{Kind: Insert, AIndex: -1, BIndex: 24, Content: "    github.com/charmbracelet/bubbletea v1.3.10"},
				{Kind: Insert, AIndex: -1, BIndex: 25, Content: "    github.com/charmbracelet/colorprofile v0.3.3 // indirect"},
				{Kind: Insert, AIndex: -1, BIndex: 26, Content: "    github.com/charmbracelet/lipgloss/v2"},
				{Kind: Insert, AIndex: -1, BIndex: 27, Content: "    github.com/charmbracelet/ultraviolet"},
				{Kind: Insert, AIndex: -1, BIndex: 28, Content: "    github.com/charmbracelet/x/ansi v0.10.3 // indirect"},
			},
			expected: []Edit{
				{Kind: Replace, AIndex: 17, BIndex: 25, Content: "    github.com/charmbracelet/colorprofile v0.3.2 // indirect", NewContent: "    github.com/charmbracelet/colorprofile v0.3.3 // indirect"},
				{Kind: Replace, AIndex: 18, BIndex: 26, Content: "    github.com/charmbracelet/lipgloss/v2", NewContent: "    github.com/charmbracelet/lipgloss/v2"},
				{Kind: Replace, AIndex: 19, BIndex: 27, Content: "    github.com/charmbracelet/ultraviolet", NewContent: "    github.com/charmbracelet/ultraviolet"},
				{Kind: Replace, AIndex: 20, BIndex: 28, Content: "    github.com/charmbracelet/x/ansi v0.10.1 // indirect", NewContent: "    github.com/charmbracelet/x/ansi v0.10.3 // indirect"},
				{Kind: Insert, AIndex: -1, BIndex: 23, Content: "    github.com/aymanbagabas/go-udiff v0.3.1 // indirect"},
				{Kind: Insert, AIndex: -1, BIndex: 24, Content: "    github.com/charmbracelet/bubbletea v1.3.10"},
			},
		},
		{
			name: "single edit",
			input: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "line1"},
			},
			expected: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "line1"},
			},
		},
		{
			name: "merge similar delete and insert",
			input: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "github.com/foo/bar v1.0.0"},
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "github.com/foo/bar v2.0.0"},
			},
			expected: []Edit{
				{Kind: Replace, AIndex: 0, BIndex: 0, Content: "github.com/foo/bar v1.0.0", NewContent: "github.com/foo/bar v2.0.0"},
			},
		},
		{
			name: "don't merge dissimilar delete and insert",
			input: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "github.com/foo/bar v1.0.0"},
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "import fmt"},
			},
			expected: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "github.com/foo/bar v1.0.0"},
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "import fmt"},
			},
		},
		{
			name: "merge insert and delete (reversed order)",
			input: []Edit{
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "github.com/foo/bar v2.0.0"},
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "github.com/foo/bar v1.0.0"},
			},
			expected: []Edit{
				{Kind: Replace, AIndex: 0, BIndex: 0, Content: "github.com/foo/bar v1.0.0", NewContent: "github.com/foo/bar v2.0.0"},
			},
		},
		{
			name: "mixed operations with merge",
			input: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "line1"},
				{Kind: Delete, AIndex: 1, BIndex: -1, Content: "github.com/foo/bar v1.0.0"},
				{Kind: Insert, AIndex: -1, BIndex: 1, Content: "github.com/foo/bar v2.0.0"},
				{Kind: Equal, AIndex: 2, BIndex: 2, Content: "line3"},
			},
			expected: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "line1"},
				{Kind: Replace, AIndex: 1, BIndex: 1, Content: "github.com/foo/bar v1.0.0", NewContent: "github.com/foo/bar v2.0.0"},
				{Kind: Equal, AIndex: 2, BIndex: 2, Content: "line3"},
			},
		},
		{
			name: "multiple inserts and deletes without merge",
			input: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "deleted line 1"},
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "new content A"},
				{Kind: Insert, AIndex: -1, BIndex: 1, Content: "new content B"},
			},
			expected: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "deleted line 1"},
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "new content A"},
				{Kind: Insert, AIndex: -1, BIndex: 1, Content: "new content B"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeReplacements(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d edits, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i].Kind != tt.expected[i].Kind {
					t.Errorf("edit %d: expected Kind %v, got %v", i, tt.expected[i].Kind, result[i].Kind)
				}
				if result[i].AIndex != tt.expected[i].AIndex {
					t.Errorf("edit %d: expected AIndex %d, got %d", i, tt.expected[i].AIndex, result[i].AIndex)
				}
				if result[i].BIndex != tt.expected[i].BIndex {
					t.Errorf("edit %d: expected BIndex %d, got %d", i, tt.expected[i].BIndex, result[i].BIndex)
				}
				if result[i].Content != tt.expected[i].Content {
					t.Errorf("edit %d: expected Content %q, got %q", i, tt.expected[i].Content, result[i].Content)
				}
				if result[i].NewContent != tt.expected[i].NewContent {
					t.Errorf("edit %d: expected NewContent %q, got %q", i, tt.expected[i].NewContent, result[i].NewContent)
				}
			}
		})
	}
}
