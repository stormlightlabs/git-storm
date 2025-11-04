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

// Parse parses a conventional commit message into structured metadata.
// Format: type(scope): description or type(scope)!: description
// Breaking changes can also be indicated by BREAKING CHANGE: in footer.
func (p *ConventionalParser) Parse(hash, subject, body string, date time.Time) (CommitMeta, error) {
	meta := CommitMeta{
		Footers: make(map[string]string),
	}

	rest := subject

	colonIdx := -1
	for i := 0; i < len(rest); i++ {
		if rest[i] == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 {
		return CommitMeta{
			Type:        "unknown",
			Description: subject,
			Body:        body,
		}, nil
	}

	prefix := rest[:colonIdx]
	description := ""
	if colonIdx+1 < len(rest) {
		description = rest[colonIdx+1:]
		if len(description) > 0 && description[0] == ' ' {
			description = description[1:]
		}
	}

	breaking := false
	if len(prefix) > 0 && prefix[len(prefix)-1] == '!' {
		breaking = true
		prefix = prefix[:len(prefix)-1]
	}

	scope := ""
	commitType := prefix

	parenStart := -1
	for i := 0; i < len(prefix); i++ {
		if prefix[i] == '(' {
			parenStart = i
			break
		}
	}

	if parenStart != -1 {
		commitType = prefix[:parenStart]
		parenEnd := -1
		for i := parenStart + 1; i < len(prefix); i++ {
			if prefix[i] == ')' {
				parenEnd = i
				break
			}
		}
		if parenEnd != -1 {
			scope = prefix[parenStart+1 : parenEnd]
		}
	}

	meta.Type = commitType
	meta.Scope = scope
	meta.Description = description
	meta.Breaking = breaking
	meta.Body = body

	if body != "" {
		lines := splitLines(body)
		inFooter := false
		currentFooter := ""
		currentValue := ""

		for _, line := range lines {
			if len(line) > 0 && !inFooter {
				colonIdx := -1
				for i := 0; i < len(line); i++ {
					if line[i] == ':' {
						colonIdx = i
						break
					}
				}
				if colonIdx != -1 {
					key := line[:colonIdx]
					value := ""
					if colonIdx+1 < len(line) {
						value = line[colonIdx+1:]
						if len(value) > 0 && value[0] == ' ' {
							value = value[1:]
						}
					}

					if key == "BREAKING CHANGE" || key == "BREAKING-CHANGE" {
						meta.Breaking = true
						inFooter = true
						currentFooter = key
						currentValue = value
						continue
					}
				}
			}

			if inFooter {
				if line == "" {
					if currentFooter != "" {
						meta.Footers[currentFooter] = currentValue
					}
					inFooter = false
					currentFooter = ""
					currentValue = ""
				} else {
					if currentValue != "" {
						currentValue += "\n"
					}
					currentValue += line
				}
			}
		}

		if inFooter && currentFooter != "" {
			meta.Footers[currentFooter] = currentValue
		}
	}

	return meta, nil
}

// IsValidType returns true if the given CommitKind is a valid conventional commit type.
func (p *ConventionalParser) IsValidType(kind CommitKind) bool {
	return kind != CommitTypeUnknown
}

// Categorize maps a CommitMeta to a changelog category.
func (p *ConventionalParser) Categorize(meta CommitMeta) string {
	switch meta.Type {
	case "feat":
		return "added"
	case "fix":
		return "fixed"
	case "perf", "refactor":
		return "changed"
	case "docs", "style", "test", "build", "ci", "chore":
		return "changed"
	case "revert":
		return "" // Skip reverts
	default:
		return "" // Unknown types are skipped
	}
}

// splitLines splits a string into lines, handling both \n and \r\n.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
