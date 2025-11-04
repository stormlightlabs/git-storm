package gitlog

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v6"
)

// CommitKind represents the kind of commit according to Conventional Commits.
type CommitKind int

const (
	CommitTypeUnknown CommitKind = iota
	CommitTypeFeat
	CommitTypeFix
	CommitTypeDocs
	CommitTypeStyle
	CommitTypeRefactor
	CommitTypePerf
	CommitTypeTest
	CommitTypeBuild
	CommitTypeCI
	CommitTypeChore
	CommitTypeRevert
)

// String returns the string representation of the CommitType.
func (kind CommitKind) String() string {
	switch kind {
	case CommitTypeFeat:
		return "feat"
	case CommitTypeFix:
		return "fix"
	case CommitTypeDocs:
		return "docs"
	case CommitTypeStyle:
		return "style"
	case CommitTypeRefactor:
		return "refactor"
	case CommitTypePerf:
		return "perf"
	case CommitTypeTest:
		return "test"
	case CommitTypeBuild:
		return "build"
	case CommitTypeCI:
		return "ci"
	case CommitTypeChore:
		return "chore"
	case CommitTypeRevert:
		return "revert"
	default:
		return "unknown"
	}
}

func Log() { fmt.Println(git.GitDirName) }

type CommitMeta struct {
	Type        string // feat, fix, docs, etc.
	Scope       string // optional
	Description string
	Breaking    bool
	Body        string
	Footers     map[string]string
}

// CommitParser defines parsing of raw commit message strings into structured metadata.
type CommitParser interface {
	// Parse takes a commit hash, subject line, body (including footers)
	// and the commit date, and returns a structured [CommitMeta]
	Parse(hash, subject, body string, date time.Time) (CommitMeta, error)

	// IsValidType returns true if the given [CommitKind] is recognised / allowed by your tooling.
	IsValidType(kind CommitKind) bool

	// Categorize returns the category (e.g., "Added", "Fixed", "Changed") for the given CommitMeta.
	Categorize(meta CommitMeta) string
}

// DefaultParser implements [CommitParser] and parses single
// or multi-line commits into one or more [CommitMeta]
type DefaultParser struct{}

// ConventionalParser implements [CommitParser] and parses
// conventional commits into one or more [CommitMeta]
type ConventionalParser struct{}
