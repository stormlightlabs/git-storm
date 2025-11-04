package gitlog

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
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

// ParseRefArgs parses command arguments to extract from/to refs.
// Supports both "from..to" and "from to" syntax.
// If only one arg, treats it as from with to=HEAD.
func ParseRefArgs(args []string) (from, to string) {
	if len(args) == 0 {
		return "", ""
	}
	if len(args) == 1 {
		parts := strings.Split(args[0], "..")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return args[0], "HEAD"
	}
	return args[0], args[1]
}

// GetCommitRange returns commits reachable from toRef but not from fromRef.
// This implements git log from..to range semantics.
func GetCommitRange(repo *git.Repository, fromRef, toRef string) ([]*object.Commit, error) {
	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", fromRef, err)
	}

	toHash, err := repo.ResolveRevision(plumbing.Revision(toRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", toRef, err)
	}

	toCommits := make(map[plumbing.Hash]bool)
	toIter, err := repo.Log(&git.LogOptions{From: *toHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", toRef, err)
	}

	err = toIter.ForEach(func(c *object.Commit) error {
		toCommits[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits from %s: %w", toRef, err)
	}

	fromCommits := make(map[plumbing.Hash]bool)
	fromIter, err := repo.Log(&git.LogOptions{From: *fromHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", fromRef, err)
	}

	err = fromIter.ForEach(func(c *object.Commit) error {
		fromCommits[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits from %s: %w", fromRef, err)
	}

	// Collect commits that are in toCommits but not in fromCommits
	result := []*object.Commit{}
	toIter, err = repo.Log(&git.LogOptions{From: *toHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", toRef, err)
	}

	err = toIter.ForEach(func(c *object.Commit) error {
		if !fromCommits[c.Hash] {
			result = append(result, c)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to collect commit range: %w", err)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// GetFileContent reads the content of a file at a specific ref (commit, tag, or branch).
func GetFileContent(repo *git.Repository, ref, filePath string) (string, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", ref, err)
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	file, err := tree.File(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	content, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// GetChangedFiles returns the list of files that changed between two commits.
func GetChangedFiles(repo *git.Repository, fromRef, toRef string) ([]string, error) {
	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", fromRef, err)
	}

	toHash, err := repo.ResolveRevision(plumbing.Revision(toRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", toRef, err)
	}

	fromCommit, err := repo.CommitObject(*fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", fromRef, err)
	}

	toCommit, err := repo.CommitObject(*toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", toRef, err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for %s: %w", fromRef, err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for %s: %w", toRef, err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	files := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.To.Name != "" {
			files = append(files, change.To.Name)
		} else {
			files = append(files, change.From.Name)
		}
	}

	return files, nil
}
