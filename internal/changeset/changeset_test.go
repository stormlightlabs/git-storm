package changeset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stormlightlabs/git-storm/internal/testutils"
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

func TestComputeDiffHash_Stability(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	commits := testutils.GetCommitHistory(t, repo)

	if len(commits) == 0 {
		t.Fatal("Expected at least one commit in test repo")
	}

	commit := commits[0]
	hash1, err := ComputeDiffHash(commit)
	if err != nil {
		t.Fatalf("ComputeDiffHash() error = %v", err)
	}

	hash2, err := ComputeDiffHash(commit)
	if err != nil {
		t.Fatalf("ComputeDiffHash() second call error = %v", err)
	}

	testutils.Expect.Equal(t, hash1, hash2, "Diff hash should be stable across multiple calls")
	testutils.Expect.Equal(t, len(hash1), 64, "Diff hash should be 64 characters (SHA256 hex)")
}

func TestComputeDiffHash_DifferentCommits(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	testutils.AddCommit(t, repo, "file1.txt", "content1", "Add file1")
	testutils.AddCommit(t, repo, "file2.txt", "content2", "Add file2")

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) < 2 {
		t.Fatal("Expected at least 2 commits")
	}

	hash1, err := ComputeDiffHash(commits[0])
	if err != nil {
		t.Fatalf("ComputeDiffHash() for commit 1 error = %v", err)
	}

	hash2, err := ComputeDiffHash(commits[1])
	if err != nil {
		t.Fatalf("ComputeDiffHash() for commit 2 error = %v", err)
	}

	testutils.Expect.NotEqual(t, hash1, hash2, "Different commits should have different diff hashes")
}

func TestWriteWithMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
		CommitHash: "abc123def456",
		DiffHash:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Type:       "added",
		Scope:      "cli",
		Summary:    "Add new feature",
		Breaking:   false,
		Author:     "Test User",
		Date:       time.Now(),
		Filename:   "",
	}

	filePath, err := WriteWithMetadata(tmpDir, meta)
	if err != nil {
		t.Fatalf("WriteWithMetadata() error = %v", err)
	}

	testutils.Expect.True(t, strings.HasSuffix(filePath, ".md"), "File path should have .md extension")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Markdown file was not created: %s", filePath)
	}

	filename := filepath.Base(filePath)
	testutils.Expect.True(t, strings.HasPrefix(filename, meta.DiffHash[:7]), "Filename should start with first 7 chars of diff hash")

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read markdown file: %v", err)
	}

	var parsedEntry Entry
	parts := strings.SplitN(string(content), "---\n", 3)
	if len(parts) < 3 {
		t.Fatal("Invalid YAML frontmatter format")
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &parsedEntry); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	testutils.Expect.Equal(t, parsedEntry.Type, meta.Type)
	testutils.Expect.Equal(t, parsedEntry.Summary, meta.Summary)
	testutils.Expect.Equal(t, parsedEntry.CommitHash, meta.CommitHash)
	testutils.Expect.Equal(t, parsedEntry.DiffHash, meta.DiffHash)

	jsonPath := filepath.Join(tmpDir, "data", meta.DiffHash+".json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("JSON metadata file was not created: %s", jsonPath)
	}

	jsonContent, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Failed to read JSON metadata: %v", err)
	}

	var parsedMeta Metadata
	if err := json.Unmarshal(jsonContent, &parsedMeta); err != nil {
		t.Fatalf("Failed to parse JSON metadata: %v", err)
	}

	testutils.Expect.Equal(t, parsedMeta.CommitHash, meta.CommitHash)
	testutils.Expect.Equal(t, parsedMeta.DiffHash, meta.DiffHash)
	testutils.Expect.Equal(t, parsedMeta.Type, meta.Type)
	testutils.Expect.Equal(t, parsedMeta.Summary, meta.Summary)
}

func TestLoadExistingMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	meta1 := Metadata{
		CommitHash: "abc123",
		DiffHash:   "hash1111111111111111111111111111111111111111111111111111111111111",
		Type:       "added",
		Summary:    "Feature 1",
		Author:     "User1",
		Date:       time.Now(),
	}

	meta2 := Metadata{
		CommitHash: "def456",
		DiffHash:   "hash2222222222222222222222222222222222222222222222222222222222222",
		Type:       "fixed",
		Summary:    "Fix 1",
		Author:     "User2",
		Date:       time.Now(),
	}

	_, err := WriteWithMetadata(tmpDir, meta1)
	if err != nil {
		t.Fatalf("Failed to write meta1: %v", err)
	}

	_, err = WriteWithMetadata(tmpDir, meta2)
	if err != nil {
		t.Fatalf("Failed to write meta2: %v", err)
	}

	loaded, err := LoadExistingMetadata(tmpDir)
	if err != nil {
		t.Fatalf("LoadExistingMetadata() error = %v", err)
	}

	testutils.Expect.Equal(t, len(loaded), 2, "Should load 2 metadata entries")

	if m, exists := loaded[meta1.DiffHash]; exists {
		testutils.Expect.Equal(t, m.CommitHash, meta1.CommitHash)
		testutils.Expect.Equal(t, m.Type, meta1.Type)
	} else {
		t.Errorf("meta1 not found in loaded metadata")
	}

	if m, exists := loaded[meta2.DiffHash]; exists {
		testutils.Expect.Equal(t, m.CommitHash, meta2.CommitHash)
		testutils.Expect.Equal(t, m.Type, meta2.Type)
	} else {
		t.Errorf("meta2 not found in loaded metadata")
	}
}

func TestLoadExistingMetadata_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	loaded, err := LoadExistingMetadata(tmpDir)
	if err != nil {
		t.Fatalf("LoadExistingMetadata() error = %v", err)
	}

	testutils.Expect.Equal(t, len(loaded), 0, "Should return empty map for non-existent data directory")
}

func TestUpdateMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
		CommitHash: "original123",
		DiffHash:   "diffhash111111111111111111111111111111111111111111111111111111111",
		Type:       "added",
		Summary:    "Feature",
		Author:     "User",
		Date:       time.Now(),
	}

	_, err := WriteWithMetadata(tmpDir, meta)
	if err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	newCommitHash := "rebased456"
	err = UpdateMetadata(tmpDir, meta.DiffHash, newCommitHash)
	if err != nil {
		t.Fatalf("UpdateMetadata() error = %v", err)
	}

	loaded, err := LoadExistingMetadata(tmpDir)
	if err != nil {
		t.Fatalf("LoadExistingMetadata() error = %v", err)
	}

	updated, exists := loaded[meta.DiffHash]
	if !exists {
		t.Fatal("Updated metadata not found")
	}

	testutils.Expect.Equal(t, updated.CommitHash, newCommitHash, "CommitHash should be updated")
	testutils.Expect.Equal(t, updated.Type, meta.Type, "Other fields should remain unchanged")
	testutils.Expect.Equal(t, updated.Summary, meta.Summary, "Other fields should remain unchanged")
}

func TestDeduplication_SameCommit(t *testing.T) {
	tmpDir := t.TempDir()
	repo := testutils.SetupTestRepo(t)
	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) == 0 {
		t.Fatal("Expected at least one commit")
	}

	commit := commits[0]
	diffHash, err := ComputeDiffHash(commit)
	if err != nil {
		t.Fatalf("ComputeDiffHash() error = %v", err)
	}

	meta := Metadata{
		CommitHash: commit.Hash.String(),
		DiffHash:   diffHash,
		Type:       "added",
		Summary:    "Test feature",
		Author:     commit.Author.Name,
		Date:       commit.Author.When,
	}

	_, err = WriteWithMetadata(tmpDir, meta)
	if err != nil {
		t.Fatalf("First WriteWithMetadata() error = %v", err)
	}

	existing, err := LoadExistingMetadata(tmpDir)
	if err != nil {
		t.Fatalf("LoadExistingMetadata() error = %v", err)
	}

	if existingMeta, exists := existing[diffHash]; exists {
		testutils.Expect.Equal(t, existingMeta.CommitHash, commit.Hash.String(), "Should detect exact duplicate")
	} else {
		t.Error("Metadata should exist in loaded entries")
	}
}

func TestDeduplication_RebasedCommit(t *testing.T) {
	tmpDir := t.TempDir()
	repo := testutils.SetupTestRepo(t)

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) == 0 {
		t.Fatal("Expected at least one commit")
	}

	commit := commits[0]
	diffHash, err := ComputeDiffHash(commit)
	if err != nil {
		t.Fatalf("ComputeDiffHash() error = %v", err)
	}

	originalMeta := Metadata{
		CommitHash: "original_commit_hash_123",
		DiffHash:   diffHash,
		Type:       "added",
		Summary:    "Test feature",
		Author:     commit.Author.Name,
		Date:       commit.Author.When,
	}

	_, err = WriteWithMetadata(tmpDir, originalMeta)
	if err != nil {
		t.Fatalf("WriteWithMetadata() error = %v", err)
	}

	existing, err := LoadExistingMetadata(tmpDir)
	if err != nil {
		t.Fatalf("LoadExistingMetadata() error = %v", err)
	}

	if existingMeta, exists := existing[diffHash]; exists {
		if existingMeta.CommitHash != commit.Hash.String() {
			err = UpdateMetadata(tmpDir, diffHash, commit.Hash.String())
			if err != nil {
				t.Fatalf("UpdateMetadata() error = %v", err)
			}

			updated, err := LoadExistingMetadata(tmpDir)
			if err != nil {
				t.Fatalf("LoadExistingMetadata() after update error = %v", err)
			}

			updatedMeta := updated[diffHash]
			testutils.Expect.Equal(t, updatedMeta.CommitHash, commit.Hash.String(), "CommitHash should be updated for rebased commit")
		}
	}
}

func TestWritePartial(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		entry       Entry
		wantErr     bool
		wantType    string
		wantSummary string
	}{
		{
			name:     "basic partial entry",
			filename: "abc1234.added.md",
			entry: Entry{
				Type:       "added",
				Scope:      "cli",
				Summary:    "Add feature",
				CommitHash: "abc123def456",
			},
			wantErr:     false,
			wantType:    "added",
			wantSummary: "Add feature",
		},
		{
			name:     "partial with different type",
			filename: "def5678.fixed.md",
			entry: Entry{
				Type:       "fixed",
				Summary:    "Fix bug",
				CommitHash: "def5678abc",
			},
			wantErr:     false,
			wantType:    "fixed",
			wantSummary: "Fix bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := WritePartial(tmpDir, tt.filename, tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("WritePartial() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			expectedPath := filepath.Join(tmpDir, tt.filename)
			testutils.Expect.Equal(t, filePath, expectedPath, "File path should match expected")

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("File was not created: %s", filePath)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			parts := strings.SplitN(string(content), "---\n", 3)
			if len(parts) < 3 {
				t.Fatal("Invalid YAML frontmatter format")
			}

			var parsed Entry
			if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			testutils.Expect.Equal(t, parsed.Type, tt.wantType)
			testutils.Expect.Equal(t, parsed.Summary, tt.wantSummary)
			testutils.Expect.Equal(t, parsed.CommitHash, tt.entry.CommitHash)
		})
	}
}

func TestWritePartial_DuplicateFilename(t *testing.T) {
	tmpDir := t.TempDir()

	filename := "abc1234.added.md"
	entry := Entry{
		Type:       "added",
		Summary:    "Test feature",
		CommitHash: "abc1234",
	}

	_, err := WritePartial(tmpDir, filename, entry)
	if err != nil {
		t.Fatalf("First WritePartial() error = %v", err)
	}

	_, err = WritePartial(tmpDir, filename, entry)
	if err == nil {
		t.Error("Expected error when writing duplicate filename, got nil")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()

	entry := Entry{
		Type:    "added",
		Scope:   "test",
		Summary: "Test deletion",
	}

	filePath, err := Write(tmpDir, entry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	filename := filepath.Base(filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File should exist before deletion: %s", filePath)
	}

	err = Delete(tmpDir, filename)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File should not exist after deletion: %s", filePath)
	}
}

func TestDelete_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := Delete(tmpDir, "nonexistent.md")
	if err == nil {
		t.Error("Expected error when deleting non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	originalEntry := Entry{
		Type:    "added",
		Scope:   "cli",
		Summary: "Original summary",
	}

	filePath, err := Write(tmpDir, originalEntry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	filename := filepath.Base(filePath)

	updatedEntry := Entry{
		Type:    "changed",
		Scope:   "api",
		Summary: "Updated summary",
	}

	err = Update(tmpDir, filename, updatedEntry)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	parts := strings.SplitN(string(content), "---\n", 3)
	if len(parts) < 3 {
		t.Fatal("Invalid YAML frontmatter format")
	}

	var parsed Entry
	if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	testutils.Expect.Equal(t, parsed.Type, updatedEntry.Type, "Type should be updated")
	testutils.Expect.Equal(t, parsed.Scope, updatedEntry.Scope, "Scope should be updated")
	testutils.Expect.Equal(t, parsed.Summary, updatedEntry.Summary, "Summary should be updated")
}

func TestUpdate_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	entry := Entry{
		Type:    "added",
		Summary: "Test",
	}

	err := Update(tmpDir, "nonexistent.md", entry)
	if err == nil {
		t.Error("Expected error when updating non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestUpdate_PreserveMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	originalEntry := Entry{
		Type:       "added",
		Scope:      "cli",
		Summary:    "Original",
		Breaking:   false,
		CommitHash: "abc123",
		DiffHash:   "def456",
	}

	filePath, err := Write(tmpDir, originalEntry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	filename := filepath.Base(filePath)

	updatedEntry := Entry{
		Type:       "changed",
		Scope:      "api",
		Summary:    "Updated",
		Breaking:   true,
		CommitHash: "abc123",
		DiffHash:   "def456",
	}

	err = Update(tmpDir, filename, updatedEntry)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	parts := strings.SplitN(string(content), "---\n", 3)
	var parsed Entry
	if err := yaml.Unmarshal([]byte(parts[1]), &parsed); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	testutils.Expect.Equal(t, parsed.CommitHash, updatedEntry.CommitHash, "CommitHash should be preserved")
	testutils.Expect.Equal(t, parsed.DiffHash, updatedEntry.DiffHash, "DiffHash should be preserved")
	testutils.Expect.Equal(t, parsed.Breaking, updatedEntry.Breaking, "Breaking should be updated")
}
