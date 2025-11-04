package changeset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		entry       Entry
		wantType    string
		wantScope   string
		wantSummary string
	}{
		{
			name: "basic entry",
			entry: Entry{
				Type:     "added",
				Scope:    "cli",
				Summary:  "Add changelog command",
				Breaking: false,
			},
			wantType:    "added",
			wantScope:   "cli",
			wantSummary: "Add changelog command",
		},
		{
			name: "entry without scope",
			entry: Entry{
				Type:     "fixed",
				Scope:    "",
				Summary:  "Fix bug in parser",
				Breaking: false,
			},
			wantType:    "fixed",
			wantScope:   "",
			wantSummary: "Fix bug in parser",
		},
		{
			name: "breaking change",
			entry: Entry{
				Type:     "changed",
				Scope:    "api",
				Summary:  "Remove legacy endpoints",
				Breaking: true,
			},
			wantType:    "changed",
			wantScope:   "api",
			wantSummary: "Remove legacy endpoints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := Write(tmpDir, tt.entry)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("File was not created: %s", filePath)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			contentStr := string(content)
			if !strings.HasPrefix(contentStr, "---\n") {
				t.Errorf("File should start with YAML frontmatter delimiter")
			}

			parts := strings.SplitN(contentStr, "---\n", 3)
			if len(parts) < 3 {
				t.Fatalf("Invalid YAML frontmatter format")
			}

			var parsed Entry
			if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", parsed.Type, tt.wantType)
			}
			if parsed.Scope != tt.wantScope {
				t.Errorf("Scope = %v, want %v", parsed.Scope, tt.wantScope)
			}
			if parsed.Summary != tt.wantSummary {
				t.Errorf("Summary = %v, want %v", parsed.Summary, tt.wantSummary)
			}
			if parsed.Breaking != tt.entry.Breaking {
				t.Errorf("Breaking = %v, want %v", parsed.Breaking, tt.entry.Breaking)
			}
		})
	}
}

func TestWrite_CollisionHandling(t *testing.T) {
	tmpDir := t.TempDir()

	entry := Entry{
		Type:    "added",
		Scope:   "test",
		Summary: "Test collision handling",
	}

	path1, err := Write(tmpDir, entry)
	if err != nil {
		t.Fatalf("First Write() error = %v", err)
	}

	path2, err := Write(tmpDir, entry)
	if err != nil {
		t.Fatalf("Second Write() error = %v", err)
	}

	if path1 == path2 {
		t.Errorf("Expected different file paths for collision, got same path: %s", path1)
	}

	if _, err := os.Stat(path1); os.IsNotExist(err) {
		t.Errorf("First file was not created: %s", path1)
	}
	if _, err := os.Stat(path2); os.IsNotExist(err) {
		t.Errorf("Second file was not created: %s", path2)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple text",
			input: "Add new feature",
			want:  "add-new-feature",
		},
		{
			name:  "text with special chars",
			input: "Fix: bug in parser!",
			want:  "fix-bug-in-parser",
		},
		{
			name:  "text with numbers",
			input: "Update version 1.2.3",
			want:  "update-version-1-2-3",
		},
		{
			name:  "text with underscores",
			input: "Add user_profile field",
			want:  "add-user-profile-field",
		},
		{
			name:  "long text gets truncated",
			input: "This is a very long summary that should be truncated to fifty characters maximum",
			want:  "this-is-a-very-long-summary-that-should-be-truncat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, "nested", "changes")

	entry := Entry{
		Type:    "added",
		Summary: "Test directory creation",
	}

	filePath, err := Write(changesDir, entry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if _, err := os.Stat(changesDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", changesDir)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File was not created: %s", filePath)
	}
}
