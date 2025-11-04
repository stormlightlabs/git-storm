package changeset

import (
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
		// File exists, add counter
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

// slugify converts a string into a URL-friendly slug.
// Converts to lowercase, replaces spaces and special chars with hyphens.
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
