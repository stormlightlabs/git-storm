package main

import (
	"strings"
	"testing"

	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestParseRefArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedFrom string
		expectedTo   string
	}{
		{
			name:         "range syntax with full hashes",
			args:         []string{"abc123..def456"},
			expectedFrom: "abc123",
			expectedTo:   "def456",
		},
		{
			name:         "range syntax with truncated hashes",
			args:         []string{"7de6f6d..18363c2"},
			expectedFrom: "7de6f6d",
			expectedTo:   "18363c2",
		},
		{
			name:         "range syntax with tags",
			args:         []string{"v1.0.0..v2.0.0"},
			expectedFrom: "v1.0.0",
			expectedTo:   "v2.0.0",
		},
		{
			name:         "two separate arguments",
			args:         []string{"abc123", "def456"},
			expectedFrom: "abc123",
			expectedTo:   "def456",
		},
		{
			name:         "single argument compares with HEAD",
			args:         []string{"abc123"},
			expectedFrom: "abc123",
			expectedTo:   "HEAD",
		},
		{
			name:         "branch names",
			args:         []string{"main", "feature-branch"},
			expectedFrom: "main",
			expectedTo:   "feature-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := parseRefArgs(tt.args)

			if from != tt.expectedFrom {
				t.Errorf("parseRefArgs() from = %v, want %v", from, tt.expectedFrom)
			}
			if to != tt.expectedTo {
				t.Errorf("parseRefArgs() to = %v, want %v", to, tt.expectedTo)
			}
		})
	}
}

func TestGetChangedFiles(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	commits := testutils.GetCommitHistory(t, repo)

	if len(commits) < 2 {
		t.Fatal("Test repo should have at least 2 commits")
	}

	fromHash := commits[1].Hash.String()
	toHash := commits[0].Hash.String()

	files, err := getChangedFiles(repo, fromHash, toHash)
	if err != nil {
		t.Fatalf("getChangedFiles() error = %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected at least one changed file")
	}

	for _, file := range files {
		if file == "" {
			t.Error("File path should not be empty")
		}
	}
}

func TestGetChangedFiles_NoChanges(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) == 0 {
		t.Fatal("Test repo should have at least 1 commit")
	}

	hash := commits[0].Hash.String()

	files, err := getChangedFiles(repo, hash, hash)
	if err != nil {
		t.Fatalf("getChangedFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected no changed files when comparing commit with itself, got %d", len(files))
	}
}

func TestGetFileContent(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) == 0 {
		t.Fatal("Test repo should have at least 1 commit")
	}

	hash := commits[0].Hash.String()

	content, err := getFileContent(repo, hash, "README.md")
	if err != nil {
		t.Fatalf("getFileContent() error = %v", err)
	}

	if content == "" {
		t.Error("Expected non-empty content for README.md")
	}

	if !strings.Contains(content, "Project") {
		t.Error("README.md should contain 'Project'")
	}
}

func TestGetFileContent_FileNotFound(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) == 0 {
		t.Fatal("Test repo should have at least 1 commit")
	}

	hash := commits[0].Hash.String()

	_, err := getFileContent(repo, hash, "nonexistent.txt")
	if err == nil {
		t.Error("Expected error when reading nonexistent file")
	}
}

func TestGetFileContent_InvalidRef(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	_, err := getFileContent(repo, "invalid-ref-12345", "README.md")
	if err == nil {
		t.Error("Expected error when using invalid ref")
	}
}
