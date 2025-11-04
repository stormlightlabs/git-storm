package gitlog

import (
	"testing"
	"time"
)

func TestConventionalParser_Parse(t *testing.T) {
	parser := &ConventionalParser{}
	testTime := time.Now()

	tests := []struct {
		name        string
		subject     string
		body        string
		wantType    string
		wantScope   string
		wantDesc    string
		wantBreak   bool
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
		name     string
		meta     CommitMeta
		wantCat  string
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
