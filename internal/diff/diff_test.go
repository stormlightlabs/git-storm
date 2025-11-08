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

func TestDiff_Compute_Unicode(t *testing.T) {
	a := []string{"Emoji 🚀", "Regular text"}
	b := []string{"Emoji 🎉", "Regular text"}

	for _, alg := range diffAlgorithms {
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error with unicode: %v", err)
			}

			reconstructed := ApplyEdits(a, edits)
			if len(reconstructed) != len(b) {
				t.Fatalf("reconstructed length %d != expected %d", len(reconstructed), len(b))
			}
			for i := range reconstructed {
				if reconstructed[i] != b[i] {
					t.Errorf("line %d: %q != %q", i, reconstructed[i], b[i])
				}
			}
		})
	}
}

func TestDiff_Compute_VeryLongLines(t *testing.T) {
	longLine1 := strings.Repeat("a", 5000)
	longLine2 := strings.Repeat("b", 5000)
	longLine3 := strings.Repeat("c", 5000)

	a := []string{longLine1, longLine2}
	b := []string{longLine1, longLine3}

	for _, alg := range diffAlgorithms {
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error with long lines: %v", err)
			}

			reconstructed := ApplyEdits(a, edits)
			if len(reconstructed) != len(b) {
				t.Fatalf("reconstructed length %d != expected %d", len(reconstructed), len(b))
			}
			for i := range reconstructed {
				if reconstructed[i] != b[i] {
					t.Errorf("line %d: lengths %d != %d", i, len(reconstructed[i]), len(b[i]))
				}
			}
		})
	}
}

func TestDiff_Compute_WhitespaceOnly(t *testing.T) {
	a := []string{"line1", "  ", "line3"}
	b := []string{"line1", "    ", "line3"}

	for _, alg := range diffAlgorithms {
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			reconstructed := ApplyEdits(a, edits)
			if len(reconstructed) != len(b) {
				t.Fatalf("reconstructed length %d != expected %d", len(reconstructed), len(b))
			}
			for i := range reconstructed {
				if reconstructed[i] != b[i] {
					t.Errorf("line %d: %q != %q", i, reconstructed[i], b[i])
				}
			}
		})
	}
}

func TestDiff_Compute_AlternatingLines(t *testing.T) {
	a := []string{"a1", "a2", "a3", "a4", "a5"}
	b := []string{"b1", "b2", "b3", "b4", "b5"}

	for _, alg := range diffAlgorithms {
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			reconstructed := ApplyEdits(a, edits)
			if len(reconstructed) != len(b) {
				t.Fatalf("reconstructed length %d != expected %d", len(reconstructed), len(b))
			}
			for i := range reconstructed {
				if reconstructed[i] != b[i] {
					t.Errorf("line %d: %q != %q", i, reconstructed[i], b[i])
				}
			}
		})
	}
}

func TestDiff_CrossValidation(t *testing.T) {
	testCases := []struct {
		name string
		a    []string
		b    []string
	}{
		{"simple", []string{"a", "b", "c"}, []string{"a", "x", "c"}},
		{"complex", []string{"1", "2", "3", "4"}, []string{"1", "x", "y", "4"}},
		{"empty to content", []string{}, []string{"a", "b", "c"}},
		{"content to empty", []string{"a", "b", "c"}, []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lcs := &LCS{}
			myers := &Myers{}

			lcsEdits, err := lcs.Compute(tc.a, tc.b)
			if err != nil {
				t.Fatalf("LCS error: %v", err)
			}

			myersEdits, err := myers.Compute(tc.a, tc.b)
			if err != nil {
				t.Fatalf("Myers error: %v", err)
			}

			lcsResult := ApplyEdits(tc.a, lcsEdits)
			myersResult := ApplyEdits(tc.a, myersEdits)

			if len(lcsResult) != len(tc.b) {
				t.Errorf("LCS reconstruction length mismatch: %d != %d", len(lcsResult), len(tc.b))
			}
			if len(myersResult) != len(tc.b) {
				t.Errorf("Myers reconstruction length mismatch: %d != %d", len(myersResult), len(tc.b))
			}

			for i := range tc.b {
				if i < len(lcsResult) && lcsResult[i] != tc.b[i] {
					t.Errorf("LCS line %d: %q != %q", i, lcsResult[i], tc.b[i])
				}
				if i < len(myersResult) && myersResult[i] != tc.b[i] {
					t.Errorf("Myers line %d: %q != %q", i, myersResult[i], tc.b[i])
				}
			}
		})
	}
}

func TestDiff_EditIndicesValid(t *testing.T) {
	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "modified", "line3", "line4"}

	for _, alg := range diffAlgorithms {
		t.Run(alg.name, func(t *testing.T) {
			m := alg.new()
			edits, err := m.Compute(a, b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i, edit := range edits {
				switch edit.Kind {
				case Equal:
					if edit.AIndex < 0 || edit.AIndex >= len(a) {
						t.Errorf("edit %d: invalid AIndex %d (len(a)=%d)", i, edit.AIndex, len(a))
					}
					if edit.BIndex < 0 || edit.BIndex >= len(b) {
						t.Errorf("edit %d: invalid BIndex %d (len(b)=%d)", i, edit.BIndex, len(b))
					}
				case Delete:
					if edit.AIndex < 0 || edit.AIndex >= len(a) {
						t.Errorf("edit %d: invalid AIndex %d for Delete", i, edit.AIndex)
					}
				case Insert:
					if edit.BIndex < 0 || edit.BIndex >= len(b) {
						t.Errorf("edit %d: invalid BIndex %d for Insert", i, edit.BIndex)
					}
				}
			}
		})
	}
}

func BenchmarkLCS_SmallInput(b *testing.B) {
	a := []string{"line1", "line2", "line3", "line4", "line5"}
	c := []string{"line1", "modified", "line3", "line4", "added"}
	lcs := &LCS{}

	for b.Loop() {
		_, _ = lcs.Compute(a, c)
	}
}

func BenchmarkMyers_SmallInput(b *testing.B) {
	a := []string{"line1", "line2", "line3", "line4", "line5"}
	c := []string{"line1", "modified", "line3", "line4", "added"}
	myers := &Myers{}

	for b.Loop() {
		_, _ = myers.Compute(a, c)
	}
}

func BenchmarkLCS_MediumInput(b *testing.B) {
	a := make([]string, 50)
	c := make([]string, 50)
	for i := range 50 {
		a[i] = "line" + strings.Repeat("x", i)
		if i%5 == 0 {
			c[i] = "modified" + strings.Repeat("y", i)
		} else {
			c[i] = a[i]
		}
	}

	lcs := &LCS{}

	for b.Loop() {
		_, _ = lcs.Compute(a, c)
	}
}

func BenchmarkMyers_MediumInput(b *testing.B) {
	a := make([]string, 50)
	c := make([]string, 50)
	for i := range 50 {
		a[i] = "line" + strings.Repeat("x", i)
		if i%5 == 0 {
			c[i] = "modified" + strings.Repeat("y", i)
		} else {
			c[i] = a[i]
		}
	}

	myers := &Myers{}

	for b.Loop() {
		_, _ = myers.Compute(a, c)
	}
}
