// TODO(determinism): Make changeset file generation deterministic using diff-based identity
//
// Current implementation uses [time.Now] for filenames, causing duplicate entries
// when generate is run multiple times on the same commit range.
//
// Store commit metadata in .changes/data/<diff-hash>.json:
// - Compute hash of git diff content (not commit message)
// - Use diff hash as stable identifier across rebases
// - Store JSON metadata: {commit_hash, diff_hash, type, scope, summary, breaking, author, date}
// - Generate .changes/<diff-hash-7>-<slug>.md from metadata
//
// Implementation:
// 1. Add DiffHash field to Entry struct
// 2. Add CommitHash field for tracking source (optional, for reference)
// 3. Create ComputeDiffHash(commit) function:
//   - Get commit.Tree() and parent.Tree()
//   - Compute diff between trees
//   - Hash the diff content (files changed + line changes)
//   - Return hex string
//
// 4. Update Write() to:
//   - Accept diff hash as parameter
//   - Use format: .changes/<diff-hash-7>-<slug>.md
//   - Write JSON to .changes/data/<diff-hash>.json
//   - Check if diff hash exists before writing (deduplication)
//
// 5. Add Read() function to parse existing entries by diff hash
//
// Directory structure:
//
//	.changes/
//	  a1b2c3d-add-authentication.md      # Human-readable entry
//	  e5f6a7b-fix-memory-leak.md
//	  data/
//	    a1b2c3d4e5f6...json              # Full metadata
//	    e5f6a7b8c9d0...json
//
// Related: See cmd/generate.go TODO for deduplication logic
package changeset

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// Entry represents a single changelog entry to be written to .changes/*.md
type Entry struct {
	Type     string `yaml:"type"`     // added, changed, fixed, removed, security
	Scope    string `yaml:"scope"`    // optional scope
	Summary  string `yaml:"summary"`  // description
	Breaking bool   `yaml:"breaking"` // true if breaking change
}

// Write creates a new .changes/<timestamp>-<slug>.md file with YAML frontmatter.
// Creates the .changes directory if it doesn't exist.
func Write(dir string, entry Entry) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	timestamp := time.Now().Format("20060102-150405")
	slug := slugify(entry.Summary)
	filename := fmt.Sprintf("%s-%s.md", timestamp, slug)
	filePath := filepath.Join(dir, filename)

	counter := 1
	for {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("%s-%s-%d.md", timestamp, slug, counter)
		filePath = filepath.Join(dir, filename)
		counter++
	}

	yamlBytes, err := yaml.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("failed to marshal entry to YAML: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n", string(yamlBytes))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return filePath, nil
}

// slugify converts a string into a URL-friendly slug by converting to lowercase,
// replaces spaces and special chars with hyphens.
func slugify(input string) string {
	s := strings.ToLower(input)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	if len(s) > 50 {
		s = s[:50]
	}

	s = strings.TrimRight(s, "-")

	return s
}

// EntryWithFile pairs an Entry with its source filename for display/processing.
type EntryWithFile struct {
	Entry    Entry
	Filename string
}

// List reads all .changes/*.md files and returns their parsed entries.
func List(dir string) ([]EntryWithFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []EntryWithFile{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var results []EntryWithFile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		parsed, err := parseEntry(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		results = append(results, EntryWithFile{
			Entry:    parsed,
			Filename: entry.Name(),
		})
	}

	return results, nil
}

// parseEntry extracts YAML frontmatter from a markdown file and unmarshals it into an Entry.
func parseEntry(content []byte) (Entry, error) {
	var entry Entry

	parts := bytes.Split(content, []byte("---"))
	if len(parts) < 3 {
		return entry, fmt.Errorf("invalid frontmatter format: expected ---...--- delimiters")
	}

	yamlContent := parts[1]
	if err := yaml.Unmarshal(yamlContent, &entry); err != nil {
		return entry, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return entry, nil
}
