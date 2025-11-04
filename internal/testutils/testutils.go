package testutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// SetupTestRepo creates a git repository in a temporary directory with sample commits.
// The repository contains multiple commits with different types of changes to support
// testing diff algorithms, changelog generation, and git log parsing.
func SetupTestRepo(t *testing.T) *git.Repository {
	t.Helper()
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create commits with varied content for diff testing
	commits := []struct {
		name, content, message string
	}{
		{"README.md", "# Project\n\nInitial version", "chore: initial commit"},
		{"a.txt", "hello world", "feat: add hello world"},
		{"b.txt", "fixed bug", "fix: patch file"},
		{"a.txt", "hello world\ngoodbye world", "feat: add goodbye"},
		{"c.txt", "new feature\nwith multiple lines\nof content", "feat: add multi-line file"},
		{"b.txt", "fixed bug\nwith proper handling", "fix: improve error handling"},
	}

	for _, c := range commits {
		path := filepath.Join(dir, c.name)
		if err := os.WriteFile(path, []byte(c.content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", c.name, err)
		}
		if _, err := w.Add(c.name); err != nil {
			t.Fatalf("failed to add file %s: %v", c.name, err)
		}
		if _, err := w.Commit(c.message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test Author",
				Email: "test@example.com",
				When:  time.Now(),
			},
		}); err != nil {
			t.Fatalf("commit failed: %v", err)
		}
	}
	return repo
}

// CreateTag creates a lightweight tag at the current HEAD of the repository.
func CreateTag(t *testing.T, repo *git.Repository, tagName string) {
	t.Helper()
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	_, err = repo.CreateTag(tagName, head.Hash(), nil)
	if err != nil {
		t.Fatalf("failed to create tag %s: %v", tagName, err)
	}
}

// GetCommitHistory returns all commits in the repository from HEAD backwards.
func GetCommitHistory(t *testing.T, repo *git.Repository) []*object.Commit {
	t.Helper()
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		t.Fatalf("failed to get commit log: %v", err)
	}

	var commits []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to iterate commits: %v", err)
	}

	return commits
}

// AddCommit adds a new commit to the repository with the given file changes.
func AddCommit(t *testing.T, repo *git.Repository, filename, content, message string) {
	t.Helper()
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	path := filepath.Join(w.Filesystem.Root(), filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", filename, err)
	}

	if _, err := w.Add(filename); err != nil {
		t.Fatalf("failed to add file %s: %v", filename, err)
	}

	if _, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
}
