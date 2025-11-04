package diff

import (
	"strings"
	"testing"
)

func TestSideBySideFormatter_Format(t *testing.T) {
	tests := []struct {
		name   string
		edits  []Edit
		width  int
		expect func(string) bool
	}{
		{
			name:  "empty edits",
			edits: []Edit{},
			width: 80,
			expect: func(output string) bool {
				return strings.Contains(output, "No changes")
			},
		},
		{
			name: "equal lines",
			edits: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "hello world"},
			},
			width: 80,
			expect: func(output string) bool {
				return strings.Contains(output, "hello world")
			},
		},
		{
			name: "insert operation",
			edits: []Edit{
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "new line"},
			},
			width: 80,
			expect: func(output string) bool {
				return strings.Contains(output, "new line") && strings.Contains(output, ">")
			},
		},
		{
			name: "delete operation",
			edits: []Edit{
				{Kind: Delete, AIndex: 0, BIndex: -1, Content: "old line"},
			},
			width: 80,
			expect: func(output string) bool {
				return strings.Contains(output, "old line") && strings.Contains(output, "<")
			},
		},
		{
			name: "mixed operations",
			edits: []Edit{
				{Kind: Equal, AIndex: 0, BIndex: 0, Content: "unchanged"},
				{Kind: Delete, AIndex: 1, BIndex: -1, Content: "removed"},
				{Kind: Insert, AIndex: -1, BIndex: 1, Content: "added"},
				{Kind: Equal, AIndex: 2, BIndex: 2, Content: "also unchanged"},
			},
			width: 100,
			expect: func(output string) bool {
				return strings.Contains(output, "unchanged") &&
					strings.Contains(output, "removed") &&
					strings.Contains(output, "added")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &SideBySideFormatter{
				TerminalWidth:   tt.width,
				ShowLineNumbers: true,
			}

			output := formatter.Format(tt.edits)

			if !tt.expect(output) {
				t.Errorf("Format() output did not meet expectations.\nGot:\n%s", output)
			}
		})
	}
}

func TestSideBySideFormatter_CalculatePaneWidth(t *testing.T) {
	tests := []struct {
		name            string
		terminalWidth   int
		showLineNumbers bool
		minExpected     int
	}{
		{
			name:            "standard width with line numbers",
			terminalWidth:   120,
			showLineNumbers: true,
			minExpected:     40,
		},
		{
			name:            "narrow terminal",
			terminalWidth:   60,
			showLineNumbers: true,
			minExpected:     minPaneWidth,
		},
		{
			name:            "without line numbers",
			terminalWidth:   100,
			showLineNumbers: false,
			minExpected:     40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &SideBySideFormatter{
				TerminalWidth:   tt.terminalWidth,
				ShowLineNumbers: tt.showLineNumbers,
			}

			paneWidth := formatter.calculatePaneWidth()

			if paneWidth < tt.minExpected {
				t.Errorf("calculatePaneWidth() = %d, expected at least %d", paneWidth, tt.minExpected)
			}
		})
	}
}

func TestSideBySideFormatter_TruncateContent(t *testing.T) {
	formatter := &SideBySideFormatter{}

	tests := []struct {
		name     string
		content  string
		maxWidth int
		expected string
	}{
		{
			name:     "short content",
			content:  "hello",
			maxWidth: 10,
			expected: "hello",
		},
		{
			name:     "exact fit",
			content:  "hello world",
			maxWidth: 11,
			expected: "hello world",
		},
		{
			name:     "needs truncation",
			content:  "hello world this is a long line",
			maxWidth: 10,
			expected: "hello w...",
		},
		{
			name:     "very small width",
			content:  "hello",
			maxWidth: 3,
			expected: "hel",
		},
		{
			name:     "trailing whitespace removed",
			content:  "hello   ",
			maxWidth: 10,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.truncateContent(tt.content, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("truncateContent() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSideBySideFormatter_RenderEdit(t *testing.T) {
	formatter := &SideBySideFormatter{
		TerminalWidth:   100,
		ShowLineNumbers: true,
	}
	paneWidth := 40

	tests := []struct {
		name   string
		edit   Edit
		expect func(left, right string) bool
	}{
		{
			name: "equal edit shows on both sides",
			edit: Edit{Kind: Equal, AIndex: 0, BIndex: 0, Content: "same"},
			expect: func(left, right string) bool {
				return strings.Contains(left, "same") && strings.Contains(right, "same")
			},
		},
		{
			name: "delete shows only on left",
			edit: Edit{Kind: Delete, AIndex: 0, BIndex: -1, Content: "removed"},
			expect: func(left, right string) bool {
				return strings.Contains(left, "removed") && !strings.Contains(right, "removed")
			},
		},
		{
			name: "insert shows only on right",
			edit: Edit{Kind: Insert, AIndex: -1, BIndex: 0, Content: "added"},
			expect: func(left, right string) bool {
				return !strings.Contains(left, "added") && strings.Contains(right, "added")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right := formatter.renderEdit(tt.edit, paneWidth)

			if !tt.expect(left, right) {
				t.Errorf("renderEdit() failed expectations.\nLeft: %q\nRight: %q", left, right)
			}
		})
	}
}
