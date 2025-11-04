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
type UnifiedDiff struct{}

// SplitDiff implements side-by-side view (old on left, new on right).
type SplitDiff struct{}

// HunkDiff focuses on changed blocks, minimal context.
type HunkDiff struct{}

// InlineDiff renders changes inline in full file flow.
type InlineDiff struct{}

// RichDiff renders changes for formatted preview (for example, markdown or html previews).
type RichDiff struct{}

// SourceDiff simply returns raw diff data or patch format without special formatting.
type SourceDiff struct{}
