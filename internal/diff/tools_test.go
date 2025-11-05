package diff

import (
	"strings"
	"testing"
)

func TestSplitDiff_Diff(t *testing.T) {
	tests := []struct {
		name        string
		oldContent  string
		newContent  string
		width       int
		showLineNum bool
		expectFunc  func(result DiffResult) bool
	}{
		{
			name:        "empty files",
			oldContent:  "",
			newContent:  "",
			width:       80,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit && strings.Contains(result.Content, "No changes")
			},
		},
		{
			name:        "identical files",
			oldContent:  "line1\nline2\nline3",
			newContent:  "line1\nline2\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "line2") &&
					strings.Contains(result.Content, "line3")
			},
		},
		{
			name:        "simple insertion",
			oldContent:  "line1\nline3",
			newContent:  "line1\nline2\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "line2") &&
					strings.Contains(result.Content, "line3") &&
					strings.Contains(result.Content, SymbolAdd)
			},
		},
		{
			name:        "simple deletion",
			oldContent:  "line1\nline2\nline3",
			newContent:  "line1\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "line2") &&
					strings.Contains(result.Content, "line3") &&
					strings.Contains(result.Content, SymbolDeleteLine)
			},
		},
		{
			name:        "replacement",
			oldContent:  "github.com/foo/bar v1.0.0",
			newContent:  "github.com/foo/bar v2.0.0",
			width:       120,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit &&
					strings.Contains(result.Content, "v1.0.0") &&
					strings.Contains(result.Content, "v2.0.0") &&
					strings.Contains(result.Content, SymbolChange)
			},
		},
		{
			name:        "without line numbers",
			oldContent:  "old line",
			newContent:  "new line",
			width:       100,
			showLineNum: false,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewSplit &&
					strings.Contains(result.Content, "old line") &&
					strings.Contains(result.Content, "new line")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := &SplitDiff{
				TerminalWidth:   tt.width,
				ShowLineNumbers: tt.showLineNum,
				Expanded:        true,
			}

			result, err := splitter.Diff(
				strings.NewReader(tt.oldContent),
				strings.NewReader(tt.newContent),
				ViewSplit,
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.expectFunc(result) {
				t.Errorf("result did not meet expectations.\nGot:\n%s", result.Content)
			}
		})
	}
}

func TestUnifiedDiff_Diff(t *testing.T) {
	tests := []struct {
		name        string
		oldContent  string
		newContent  string
		width       int
		showLineNum bool
		expectFunc  func(result DiffResult) bool
	}{
		{
			name:        "empty files",
			oldContent:  "",
			newContent:  "",
			width:       80,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified && strings.Contains(result.Content, "No changes")
			},
		},
		{
			name:        "identical files",
			oldContent:  "line1\nline2\nline3",
			newContent:  "line1\nline2\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "line2") &&
					strings.Contains(result.Content, "line3")
			},
		},
		{
			name:        "simple insertion",
			oldContent:  "line1\nline3",
			newContent:  "line1\nline2\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "+line2") &&
					strings.Contains(result.Content, "line3")
			},
		},
		{
			name:        "simple deletion",
			oldContent:  "line1\nline2\nline3",
			newContent:  "line1\nline3",
			width:       100,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified &&
					strings.Contains(result.Content, "line1") &&
					strings.Contains(result.Content, "-line2") &&
					strings.Contains(result.Content, "line3")
			},
		},
		{
			name:        "replacement",
			oldContent:  "github.com/foo/bar v1.0.0",
			newContent:  "github.com/foo/bar v2.0.0",
			width:       120,
			showLineNum: true,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified &&
					strings.Contains(result.Content, "-github.com/foo/bar v1.0.0") &&
					strings.Contains(result.Content, "+github.com/foo/bar v2.0.0")
			},
		},
		{
			name:        "without line numbers",
			oldContent:  "old line",
			newContent:  "new line",
			width:       100,
			showLineNum: false,
			expectFunc: func(result DiffResult) bool {
				return result.View == ViewUnified &&
					strings.Contains(result.Content, "-old line") &&
					strings.Contains(result.Content, "+new line")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unifier := &UnifiedDiff{
				TerminalWidth:   tt.width,
				ShowLineNumbers: tt.showLineNum,
				Expanded:        true,
			}

			result, err := unifier.Diff(
				strings.NewReader(tt.oldContent),
				strings.NewReader(tt.newContent),
				ViewUnified,
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.expectFunc(result) {
				t.Errorf("result did not meet expectations.\nGot:\n%s", result.Content)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single line no newline",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			input:    "hello\n",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multiple lines with trailing newline",
			input:    "line1\nline2\nline3\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines preserved",
			input:    "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{"", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestDiffViewKind_String(t *testing.T) {
	tests := []struct {
		kind     DiffViewKind
		expected string
	}{
		{ViewUnified, "Unified"},
		{ViewSplit, "Split"},
		{ViewHunk, "Hunk"},
		{ViewInline, "Inline"},
		{ViewRich, "Rich"},
		{ViewSource, "Source"},
		{DiffViewKind(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSplitDiff_CompressedView(t *testing.T) {
	oldLines := make([]string, 50)
	newLines := make([]string, 50)
	for i := range 50 {
		oldLines[i] = "unchanged line"
		newLines[i] = "unchanged line"
	}

	newLines[25] = "changed line"

	oldContent := strings.Join(oldLines, "\n")
	newContent := strings.Join(newLines, "\n")

	splitter := &SplitDiff{
		TerminalWidth:   100,
		ShowLineNumbers: true,
		Expanded:        false,
	}

	result, err := splitter.Diff(
		strings.NewReader(oldContent),
		strings.NewReader(newContent),
		ViewSplit,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Content, "unchanged lines") {
		t.Errorf("expected compression indicator in output")
	}

	if !strings.Contains(result.Content, "changed line") {
		t.Errorf("expected changed line in output")
	}
}

func TestUnifiedDiff_CompressedView(t *testing.T) {
	oldLines := make([]string, 50)
	newLines := make([]string, 50)
	for i := range 50 {
		oldLines[i] = "unchanged line"
		newLines[i] = "unchanged line"
	}

	newLines[25] = "changed line"

	oldContent := strings.Join(oldLines, "\n")
	newContent := strings.Join(newLines, "\n")

	unifier := &UnifiedDiff{
		TerminalWidth:   100,
		ShowLineNumbers: true,
		Expanded:        false, // Enable compression
	}

	result, err := unifier.Diff(
		strings.NewReader(oldContent),
		strings.NewReader(newContent),
		ViewUnified,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Content, "unchanged lines") {
		t.Errorf("expected compression indicator in output")
	}

	if !strings.Contains(result.Content, "+changed line") {
		t.Errorf("expected changed line with + prefix in output")
	}
}
