package diff

import (
	"strings"
	"testing"
)

func TestSideBySideFormatter_CompressUnchangedBlocks(t *testing.T) {
	tests := []struct {
		name                              string
		edits                             []Edit
		expectedCompressed, expectedTotal int
	}{
		{
			name:               "no compression for small unchanged blocks",
			edits:              makeEqualEdits(5),
			expectedCompressed: 0,
			expectedTotal:      5,
		},
		{
			name:               "compress large unchanged block",
			edits:              makeEqualEdits(20),
			expectedCompressed: 1,
			expectedTotal:      7,
		},
		{
			name: "compress unchanged between changes",
			edits: []Edit{
				{Kind: Insert, AIndex: -1, BIndex: 0, Content: "new line"},
				{Kind: Equal, AIndex: 0, BIndex: 1, Content: "unchanged 1"},
				{Kind: Equal, AIndex: 1, BIndex: 2, Content: "unchanged 2"},
				{Kind: Equal, AIndex: 2, BIndex: 3, Content: "unchanged 3"},
				{Kind: Equal, AIndex: 3, BIndex: 4, Content: "unchanged 4"},
				{Kind: Equal, AIndex: 4, BIndex: 5, Content: "unchanged 5"},
				{Kind: Equal, AIndex: 5, BIndex: 6, Content: "unchanged 6"},
				{Kind: Equal, AIndex: 6, BIndex: 7, Content: "unchanged 7"},
				{Kind: Equal, AIndex: 7, BIndex: 8, Content: "unchanged 8"},
				{Kind: Equal, AIndex: 8, BIndex: 9, Content: "unchanged 9"},
				{Kind: Equal, AIndex: 9, BIndex: 10, Content: "unchanged 10"},
				{Kind: Equal, AIndex: 10, BIndex: 11, Content: "unchanged 11"},
				{Kind: Equal, AIndex: 11, BIndex: 12, Content: "unchanged 12"},
				{Kind: Equal, AIndex: 12, BIndex: 13, Content: "unchanged 13"},
				{Kind: Delete, AIndex: 13, BIndex: -1, Content: "removed line"},
			},
			expectedCompressed: 1,
			expectedTotal:      9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &SideBySideFormatter{}

			result := formatter.compressUnchangedBlocks(tt.edits)

			if len(result) != tt.expectedTotal {
				t.Errorf("Expected %d total edits after compression, got %d", tt.expectedTotal, len(result))
			}

			compressed := countCompressedBlocks(result)
			if compressed != tt.expectedCompressed {
				t.Errorf("Expected %d compressed blocks, got %d", tt.expectedCompressed, compressed)
			}
		})
	}
}

func TestSideBySideFormatter_Expanded(t *testing.T) {
	edits := makeEqualEdits(20)

	compressedFormatter := &SideBySideFormatter{
		TerminalWidth:   100,
		ShowLineNumbers: true,
		Expanded:        false,
	}

	expandedFormatter := &SideBySideFormatter{
		TerminalWidth:   100,
		ShowLineNumbers: true,
		Expanded:        true,
	}

	compressedOutput := compressedFormatter.Format(edits)
	expandedOutput := expandedFormatter.Format(edits)

	compressedLines := strings.Split(compressedOutput, "\n")
	expandedLines := strings.Split(expandedOutput, "\n")

	if len(expandedLines) <= len(compressedLines) {
		t.Errorf("Expanded output (%d lines) should have more lines than compressed (%d lines)",
			len(expandedLines), len(compressedLines))
	}

	if !strings.Contains(compressedOutput, compressedIndicator) {
		t.Error("Compressed output should contain compression indicator")
	}

	if strings.Contains(expandedOutput, compressedIndicator) {
		t.Error("Expanded output should not contain compression indicator")
	}
}

func TestSideBySideFormatter_CompressedIndicatorStyling(t *testing.T) {
	formatter := &SideBySideFormatter{
		TerminalWidth:   100,
		ShowLineNumbers: true,
		Expanded:        false,
	}

	edits := makeEqualEdits(20)
	output := formatter.Format(edits)

	if !strings.Contains(output, "unchanged lines") {
		t.Error("Compressed output should mention 'unchanged lines'")
	}
}

func TestMultipleCompressedBlocks(t *testing.T) {
	formatter := &SideBySideFormatter{}

	edits := []Edit{}
	edits = append(edits, makeEqualEdits(20)...)
	edits = append(edits, Edit{Kind: Insert, AIndex: -1, BIndex: 0, Content: "change 1"})
	edits = append(edits, makeEqualEdits(20)...)
	edits = append(edits, Edit{Kind: Delete, AIndex: 0, BIndex: -1, Content: "change 2"})
	edits = append(edits, makeEqualEdits(20)...)

	result := formatter.compressUnchangedBlocks(edits)

	compressed := countCompressedBlocks(result)
	if compressed != 3 {
		t.Errorf("Expected 3 compressed blocks, got %d", compressed)
	}
}

func makeEqualEdits(count int) []Edit {
	edits := make([]Edit, count)
	for i := range count {
		edits[i] = Edit{
			Kind:    Equal,
			AIndex:  i,
			BIndex:  i,
			Content: "unchanged line",
		}
	}
	return edits
}

func countCompressedBlocks(edits []Edit) int {
	count := 0
	for _, edit := range edits {
		if edit.AIndex == -2 && edit.BIndex == -2 {
			count++
		}
	}
	return count
}
