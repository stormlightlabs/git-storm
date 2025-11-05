package diff

import "io"

// DiffViewKind enumerates the supported diff views.
type DiffViewKind int

const (
	ViewUnified DiffViewKind = iota
	ViewSplit
	ViewHunk
	ViewInline
	ViewRich
	ViewSource
)

// String returns a human-readable name for the view kind.
func (k DiffViewKind) String() string {
	switch k {
	case ViewUnified:
		return "Unified"
	case ViewSplit:
		return "Split"
	case ViewHunk:
		return "Hunk"
	case ViewInline:
		return "Inline"
	case ViewRich:
		return "Rich"
	case ViewSource:
		return "Source"
	default:
		return "Unknown"
	}
}

// DiffResult holds the output of a diff tool.
type DiffResult struct {
	// Content is the rendered diff output (text, ANSI, HTML, etc.)
	Content string

	// View is the kind of view used to render the diff.
	View DiffViewKind

	// Metadata like file paths, commit hashes, timestamps (optional).
	// You may extend as needed.
	OldPath  string
	NewPath  string
	FromHash string
	ToHash   string
}

// DiffTool is the interface for generating diffs between two versions.
type DiffTool interface {
	// Diff takes two blobs (old and new versions) and returns a DiffResult.
	// The reader parameters may be full file contents or other abstraction.
	// The viewKind parameter selects which view implementation (Unified/Split/...).
	Diff(oldContent io.Reader, newContent io.Reader, viewKind DiffViewKind) (DiffResult, error)
}

// UnifiedDiff implements unified view (single linear view with additions & deletions).
//
// TODO: Support pluggable diff algorithms beyond Myers.
type UnifiedDiff struct {
	// TerminalWidth is the total available width for rendering
	TerminalWidth int
	// ShowLineNumbers controls whether line numbers are displayed
	ShowLineNumbers bool
	// Expanded controls whether to show all unchanged lines or compress them
	Expanded bool
	// EnableWordWrap enables word wrapping for long lines
	EnableWordWrap bool
}

// Diff generates a unified diff view from two content readers.
func (u *UnifiedDiff) Diff(oldContent io.Reader, newContent io.Reader, viewKind DiffViewKind) (DiffResult, error) {
	oldBytes, err := io.ReadAll(oldContent)
	if err != nil {
		return DiffResult{}, err
	}
	newBytes, err := io.ReadAll(newContent)
	if err != nil {
		return DiffResult{}, err
	}

	oldLines := splitLines(string(oldBytes))
	newLines := splitLines(string(newBytes))

	myers := &Myers{}
	edits, err := myers.Compute(oldLines, newLines)
	if err != nil {
		return DiffResult{}, err
	}

	formatter := &UnifiedFormatter{
		TerminalWidth:   u.TerminalWidth,
		ShowLineNumbers: u.ShowLineNumbers,
		Expanded:        u.Expanded,
		EnableWordWrap:  u.EnableWordWrap,
	}

	content := formatter.Format(edits)

	return DiffResult{
		Content: content,
		View:    ViewUnified,
	}, nil
}

// SplitDiff implements side-by-side view (old on left, new on right).
//
// TODO: Support pluggable diff algorithms beyond Myers.
type SplitDiff struct {
	// TerminalWidth is the total available width for rendering
	TerminalWidth int
	// ShowLineNumbers controls whether line numbers are displayed
	ShowLineNumbers bool
	// Expanded controls whether to show all unchanged lines or compress them
	Expanded bool
	// EnableWordWrap enables word wrapping for long lines
	EnableWordWrap bool
}

// Diff generates a side-by-side diff view from two content readers.
func (s *SplitDiff) Diff(oldContent io.Reader, newContent io.Reader, viewKind DiffViewKind) (DiffResult, error) {
	oldBytes, err := io.ReadAll(oldContent)
	if err != nil {
		return DiffResult{}, err
	}
	newBytes, err := io.ReadAll(newContent)
	if err != nil {
		return DiffResult{}, err
	}

	oldLines := splitLines(string(oldBytes))
	newLines := splitLines(string(newBytes))

	myers := &Myers{}
	edits, err := myers.Compute(oldLines, newLines)
	if err != nil {
		return DiffResult{}, err
	}

	formatter := &SideBySideFormatter{
		TerminalWidth:   s.TerminalWidth,
		ShowLineNumbers: s.ShowLineNumbers,
		Expanded:        s.Expanded,
		EnableWordWrap:  s.EnableWordWrap,
	}

	content := formatter.Format(edits)

	return DiffResult{
		Content: content,
		View:    ViewSplit,
	}, nil
}

// splitLines splits a string into lines, preserving empty lines.
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// HunkDiff focuses on changed blocks, minimal context.
type HunkDiff struct{}

// InlineDiff renders changes inline in full file flow.
type InlineDiff struct{}

// RichDiff renders changes for formatted preview (for example, markdown or html previews).
type RichDiff struct{}

// SourceDiff simply returns raw diff data or patch format without special formatting.
type SourceDiff struct{}
