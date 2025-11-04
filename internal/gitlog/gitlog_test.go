package gitlog

import (
	"testing"
	"time"

	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestConventionalParser_Parse(t *testing.T) {
	parser := &ConventionalParser{}
	testTime := time.Now()

	tests := []struct {
		name      string
		subject   string
		body      string
		wantType  string
		wantScope string
		wantDesc  string
		wantBreak bool
	}{
		{
			name:      "simple feat",
			subject:   "feat: add new feature",
			body:      "",
			wantType:  "feat",
			wantScope: "",
			wantDesc:  "add new feature",
			wantBreak: false,
		},
		{
			name:      "feat with scope",
			subject:   "feat(api): add pagination endpoint",
			body:      "",
			wantType:  "feat",
			wantScope: "api",
			wantDesc:  "add pagination endpoint",
			wantBreak: false,
		},
		{
			name:      "fix with scope",
			subject:   "fix(ui): correct button alignment issue",
			body:      "",
			wantType:  "fix",
			wantScope: "ui",
			wantDesc:  "correct button alignment issue",
			wantBreak: false,
		},
		{
			name:      "breaking change with !",
			subject:   "feat(api)!: remove support for legacy endpoints",
			body:      "",
			wantType:  "feat",
			wantScope: "api",
			wantDesc:  "remove support for legacy endpoints",
			wantBreak: true,
		},
		{
			name:      "breaking change without scope",
			subject:   "feat!: major API redesign",
			body:      "",
			wantType:  "feat",
			wantScope: "",
			wantDesc:  "major API redesign",
			wantBreak: true,
		},
		{
			name:      "breaking change in footer",
			subject:   "feat(api): update authentication",
			body:      "Some details here\n\nBREAKING CHANGE: API no longer accepts XML-formatted requests.",
			wantType:  "feat",
			wantScope: "api",
			wantDesc:  "update authentication",
			wantBreak: true,
		},
		{
			name:      "docs commit",
			subject:   "docs: update README installation instructions",
			body:      "",
			wantType:  "docs",
			wantScope: "",
			wantDesc:  "update README installation instructions",
			wantBreak: false,
		},
		{
			name:      "chore commit",
			subject:   "chore: update .gitignore",
			body:      "",
			wantType:  "chore",
			wantScope: "",
			wantDesc:  "update .gitignore",
			wantBreak: false,
		},
		{
			name:      "non-conventional commit",
			subject:   "some random commit message",
			body:      "",
			wantType:  "unknown",
			wantScope: "",
			wantDesc:  "some random commit message",
			wantBreak: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := parser.Parse("abc123", tt.subject, tt.body, testTime)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if meta.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", meta.Type, tt.wantType)
			}
			if meta.Scope != tt.wantScope {
				t.Errorf("Scope = %v, want %v", meta.Scope, tt.wantScope)
			}
			if meta.Description != tt.wantDesc {
				t.Errorf("Description = %v, want %v", meta.Description, tt.wantDesc)
			}
			if meta.Breaking != tt.wantBreak {
				t.Errorf("Breaking = %v, want %v", meta.Breaking, tt.wantBreak)
			}
		})
	}
}

func TestConventionalParser_Categorize(t *testing.T) {
	parser := &ConventionalParser{}

	tests := []struct {
		name    string
		meta    CommitMeta
		wantCat string
	}{
		{
			name:    "feat -> added",
			meta:    CommitMeta{Type: "feat"},
			wantCat: "added",
		},
		{
			name:    "fix -> fixed",
			meta:    CommitMeta{Type: "fix"},
			wantCat: "fixed",
		},
		{
			name:    "perf -> changed",
			meta:    CommitMeta{Type: "perf"},
			wantCat: "changed",
		},
		{
			name:    "refactor -> changed",
			meta:    CommitMeta{Type: "refactor"},
			wantCat: "changed",
		},
		{
			name:    "docs -> changed",
			meta:    CommitMeta{Type: "docs"},
			wantCat: "changed",
		},
		{
			name:    "test -> changed",
			meta:    CommitMeta{Type: "test"},
			wantCat: "changed",
		},
		{
			name:    "revert -> skip",
			meta:    CommitMeta{Type: "revert"},
			wantCat: "",
		},
		{
			name:    "unknown -> skip",
			meta:    CommitMeta{Type: "unknown"},
			wantCat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Categorize(tt.meta)
			if got != tt.wantCat {
				t.Errorf("Categorize() = %v, want %v", got, tt.wantCat)
			}
		})
	}
}

func TestConventionalParser_IsValidType(t *testing.T) {
	parser := &ConventionalParser{}

	tests := []struct {
		name string
		kind CommitKind
		want bool
	}{
		{
			name: "feat is valid",
			kind: CommitTypeFeat,
			want: true,
		},
		{
			name: "fix is valid",
			kind: CommitTypeFix,
			want: true,
		},
		{
			name: "unknown is invalid",
			kind: CommitTypeUnknown,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.IsValidType(tt.kind)
			if got != tt.want {
				t.Errorf("IsValidType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRefArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFrom string
		wantTo   string
	}{
		{
			name:     "range syntax",
			args:     []string{"v1.0.0..v1.1.0"},
			wantFrom: "v1.0.0",
			wantTo:   "v1.1.0",
		},
		{
			name:     "two separate args",
			args:     []string{"v1.0.0", "v1.1.0"},
			wantFrom: "v1.0.0",
			wantTo:   "v1.1.0",
		},
		{
			name:     "single arg defaults to HEAD",
			args:     []string{"v1.0.0"},
			wantFrom: "v1.0.0",
			wantTo:   "HEAD",
		},
		{
			name:     "empty args",
			args:     []string{},
			wantFrom: "",
			wantTo:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := ParseRefArgs(tt.args)
			if from != tt.wantFrom {
				t.Errorf("ParseRefArgs() from = %v, want %v", from, tt.wantFrom)
			}
			if to != tt.wantTo {
				t.Errorf("ParseRefArgs() to = %v, want %v", to, tt.wantTo)
			}
		})
	}
}

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

	rangeCommits, err := GetCommitRange(repo, "v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("GetCommitRange() error = %v", err)
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

	rangeCommits, err := GetCommitRange(repo, "HEAD", "HEAD")
	if err != nil {
		t.Fatalf("GetCommitRange() error = %v", err)
	}

	if len(rangeCommits) != 0 {
		t.Errorf("Expected 0 commits when from and to are the same, got %d", len(rangeCommits))
	}
}

func TestGetFileContent(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	content, err := GetFileContent(repo, "HEAD", "README.md")
	if err != nil {
		t.Fatalf("GetFileContent() error = %v", err)
	}

	if content == "" {
		t.Errorf("Expected non-empty content for README.md")
	}

	if content != "# Project\n\nInitial version" {
		t.Errorf("GetFileContent() content = %v, want %v", content, "# Project\\n\\nInitial version")
	}
}

func TestGetFileContent_InvalidFile(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	_, err := GetFileContent(repo, "HEAD", "nonexistent.txt")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

func TestGetChangedFiles(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	commits := testutils.GetCommitHistory(t, repo)
	if len(commits) < 2 {
		t.Fatalf("Expected at least 2 commits, got %d", len(commits))
	}

	files, err := GetChangedFiles(repo, commits[1].Hash.String(), commits[0].Hash.String())
	if err != nil {
		t.Fatalf("GetChangedFiles() error = %v", err)
	}

	if len(files) == 0 {
		t.Errorf("Expected at least 1 changed file, got 0")
	}
}

func TestGetChangedFiles_NoChanges(t *testing.T) {
	repo := testutils.SetupTestRepo(t)

	files, err := GetChangedFiles(repo, "HEAD", "HEAD")
	if err != nil {
		t.Fatalf("GetChangedFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 changed files when refs are the same, got %d", len(files))
	}
}
