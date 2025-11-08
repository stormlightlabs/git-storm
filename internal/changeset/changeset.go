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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/goccy/go-yaml"
)

// Entry represents a single changelog entry to be written to .changes/*.md
type Entry struct {
	Type       string `yaml:"type"`                  // added, changed, fixed, removed, security
	Scope      string `yaml:"scope"`                 // optional scope
	Summary    string `yaml:"summary"`               // description
	Breaking   bool   `yaml:"breaking"`              // true if breaking change
	CommitHash string `yaml:"commit_hash,omitempty"` // source commit hash (for reference)
	DiffHash   string `yaml:"diff_hash,omitempty"`   // hash of git diff content (for deduplication)
}

// Metadata stores complete entry information in .changes/data/*.json for deduplication
type Metadata struct {
	CommitHash string    `json:"commit_hash"` // current commit hash
	DiffHash   string    `json:"diff_hash"`   // stable diff content hash
	Filename   string    `json:"filename"`    // relative path to .md file
	Type       string    `json:"type"`
	Scope      string    `json:"scope"`
	Summary    string    `json:"summary"`
	Breaking   bool      `json:"breaking"`
	Author     string    `json:"author"`
	Date       time.Time `json:"date"`
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

// WritePartial creates a .changes/<filename> file with the specified name and YAML frontmatter.
// This is used by the `unreleased partial` command to create entries with commit-hash based names.
// Creates the .changes directory if it doesn't exist.
func WritePartial(dir string, filename string, entry Entry) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("file %s already exists", filename)
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

// WriteWithMetadata creates a new .changes/<diffHash7>-<slug>.md file with YAML
// frontmatter and saves corresponding metadata to .changes/data/<diffHash>.json.
//
// The filename uses the first 7 characters of the diff hash for human-readable
// identification, while the JSON metadata file uses the full hash for
// deduplication lookups.
func WriteWithMetadata(dir string, meta Metadata) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	diffHashShort := meta.DiffHash[:7]
	slug := slugify(meta.Summary)
	filename := fmt.Sprintf("%s-%s.md", diffHashShort, slug)
	filePath := filepath.Join(dir, filename)

	entry := Entry{
		Type:       meta.Type,
		Scope:      meta.Scope,
		Summary:    meta.Summary,
		Breaking:   meta.Breaking,
		CommitHash: meta.CommitHash,
		DiffHash:   meta.DiffHash,
	}

	yamlBytes, err := yaml.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("failed to marshal entry to YAML: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n", string(yamlBytes))
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	meta.Filename = filename
	if err := SaveMetadata(dir, meta); err != nil {
		return "", fmt.Errorf("failed to save metadata: %w", err)
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

// ComputeDiffHash calculates a stable hash of the commit's diff content. This
// hash is independent of the commit hash, so rebased commits with identical
// diffs will produce the same hash.
//
// The hash is computed from:
//   - Sorted list of changed file paths
//   - For each file: the full diff content (additions and deletions)
func ComputeDiffHash(commit *object.Commit) (string, error) {
	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get commit tree: %w", err)
	}

	var parentTree *object.Tree
	if commit.NumParents() > 0 {
		parent, err := commit.Parent(0)
		if err != nil {
			return "", fmt.Errorf("failed to get parent commit: %w", err)
		}
		parentTree, err = parent.Tree()
		if err != nil {
			return "", fmt.Errorf("failed to get parent tree: %w", err)
		}
	}

	var changes object.Changes
	if parentTree != nil {
		changes, err = parentTree.Diff(tree)
		if err != nil {
			return "", fmt.Errorf("failed to compute diff: %w", err)
		}
	} else {
		emptyTree := &object.Tree{}
		changes, err = object.DiffTreeWithOptions(context.TODO(), emptyTree, tree, &object.DiffTreeOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to compute diff for initial commit: %w", err)
		}
	}

	var diffParts []string
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			return "", fmt.Errorf("failed to get patch for %s: %w", change.To.Name, err)
		}

		diffParts = append(diffParts, fmt.Sprintf("FILE:%s\n%s", change.To.Name, patch.String()))
	}

	sort.Strings(diffParts)

	hasher := sha256.New()
	for _, part := range diffParts {
		hasher.Write([]byte(part))
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// SaveMetadata writes metadata to .changes/data/<diffHash>.json
func SaveMetadata(dir string, meta Metadata) error {
	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filePath := filepath.Join(dataDir, meta.DiffHash+".json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// LoadExistingMetadata reads all metadata files from .changes/data/*.json
// and creates a map of diff hash -> metadata for O(1) lookups.
func LoadExistingMetadata(dir string) (map[string]Metadata, error) {
	dataDir := filepath.Join(dir, "data")
	result := make(map[string]Metadata)
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dataDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata file %s: %w", entry.Name(), err)
		}

		var meta Metadata
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata from %s: %w", entry.Name(), err)
		}

		result[meta.DiffHash] = meta
	}
	return result, nil
}

// UpdateMetadata updates an existing metadata file with a new commit hash when
// a rebased commit is detected (same diff, different commit hash).
func UpdateMetadata(dir string, diffHash string, newCommitHash string) error {
	dataDir := filepath.Join(dir, "data")
	filePath := filepath.Join(dataDir, diffHash+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read existing metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	meta.CommitHash = newCommitHash

	updatedData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated metadata: %w", err)
	}

	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}
	return nil
}
