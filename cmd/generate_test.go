package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestGetCommitRange(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) < 3 {
		t.Fatalf("Expected at least 3 commits, got %d", len(commits))
	}

	oldCommit := commits[len(commits)-2]
	if err := testutils.CreateTagAtCommit(t, repo, "v1.0.0", oldCommit.Hash.String()); err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	testutils.AddCommit(t, repo, "d.txt", "content d", "feat: add d feature")
	testutils.AddCommit(t, repo, "e.txt", "content e", "fix: fix e bug")

	rangeCommits, err := gitlog.GetCommitRange(repo, "v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("gitlog.GetCommitRange() error = %v", err)
	}

	if len(rangeCommits) < 2 {
		t.Errorf("Expected at least 2 commits in range, got %d", len(rangeCommits))
	}

	for i := 1; i < len(rangeCommits); i++ {
		if rangeCommits[i].Author.When.Before(rangeCommits[i-1].Author.When) {
			t.Errorf("Commits are not in chronological order")
		}
	}
}

func TestGetCommitRange_SameRef(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	rangeCommits, err := gitlog.GetCommitRange(repo, "HEAD", "HEAD")
	if err != nil {
		t.Fatalf("gitlog.GetCommitRange() error = %v", err)
	}

	if len(rangeCommits) != 0 {
		t.Errorf("Expected 0 commits when from and to are the same, got %d", len(rangeCommits))
	}
}

func TestGetCommitRange_InvalidRef(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	_, err := gitlog.GetCommitRange(repo, "invalid-ref", "HEAD")
	if err == nil {
		t.Errorf("Expected error for invalid ref, got nil")
	}
}

func TestGenerateCmd_JSONOutput(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) < 2 {
		t.Fatalf("Expected at least 2 commits, got %d", len(commits))
	}

	oldCommit := commits[len(commits)-2]
	if err := testutils.CreateTagAtCommit(t, repo, "v1.0.0", oldCommit.Hash.String()); err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	testutils.AddCommit(t, repo, "feat.txt", "content", "feat: add new feature")
	testutils.AddCommit(t, repo, "fix.txt", "content", "fix: fix bug")

	repoPath = worktree.Filesystem.Root()
	outputJSON = true

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		os.Chdir(oldWd)
		outputJSON = false
	}()

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cmd := generateCmd()
	cmd.SetArgs([]string{"v1.0.0", "HEAD"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("generateCmd() error = %v", err)
	}
}

func TestGenerateCmd_InteractiveAndJSONConflict(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	repoPath = worktree.Filesystem.Root()

	cmd := generateCmd()
	cmd.SetArgs([]string{"--interactive", "--output-json", "HEAD~1", "HEAD"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error when using --interactive and --output-json together, got nil")
	}

	if err != nil {
		validErrors := []string{
			"--interactive and --output-json cannot be used together",
			"requires an interactive terminal",
		}
		foundValidError := false
		for _, validErr := range validErrors {
			if strings.Contains(err.Error(), validErr) {
				foundValidError = true
				break
			}
		}
		if !foundValidError {
			t.Errorf("Expected error about flags conflict or TTY requirement, got: %v", err)
		}
	}
}
