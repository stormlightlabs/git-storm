package main

import (
	"strings"
	"testing"

	"github.com/stormlightlabs/git-storm/internal/changelog"
	"github.com/stormlightlabs/git-storm/internal/testutils"
)

func TestCreateReleaseTag(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	repoPath := worktree.Filesystem.Root()
	version := &changelog.Version{
		Number: "1.0.0",
		Date:   "2024-01-15",
		Sections: []changelog.Section{
			{
				Type: "added",
				Entries: []string{
					"New authentication system",
					"User profile management",
				},
			},
			{
				Type: "fixed",
				Entries: []string{
					"Memory leak in database connection pool",
				},
			},
		},
	}

	err = createReleaseTag(repoPath, "1.0.0", version)
	if err != nil {
		t.Fatalf("createReleaseTag() error = %v", err)
	}

	tagRef, err := repo.Tag("v1.0.0")
	if err != nil {
		t.Fatalf("Tag v1.0.0 should exist, got error: %v", err)
	}

	tagObj, err := repo.TagObject(tagRef.Hash())
	if err != nil {
		t.Fatalf("Tag should be annotated, got error: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Failed to get HEAD: %v", err)
	}

	testutils.Expect.Equal(t, tagObj.Target, head.Hash(), "Tag should point to HEAD")

	message := tagObj.Message
	testutils.Expect.True(t, strings.Contains(message, "Release 1.0.0"), "Tag message should contain version")
	testutils.Expect.True(t, strings.Contains(message, "Added:"), "Tag message should contain Added section")
	testutils.Expect.True(t, strings.Contains(message, "Fixed:"), "Tag message should contain Fixed section")
	testutils.Expect.True(t, strings.Contains(message, "New authentication system"), "Tag message should contain entry")
	testutils.Expect.True(t, strings.Contains(message, "Memory leak"), "Tag message should contain entry")
}

func TestCreateReleaseTag_DuplicateTag(t *testing.T) {
	repo := testutils.SetupTestRepo(t)
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	repoPath := worktree.Filesystem.Root()
	version := &changelog.Version{
		Number: "1.0.0",
		Date:   "2024-01-15",
		Sections: []changelog.Section{
			{
				Type:    "added",
				Entries: []string{"Feature 1"},
			},
		},
	}

	err = createReleaseTag(repoPath, "1.0.0", version)
	if err != nil {
		t.Fatalf("First createReleaseTag() error = %v", err)
	}

	err = createReleaseTag(repoPath, "1.0.0", version)
	if err == nil {
		t.Error("Expected error when creating duplicate tag, got nil")
	}

	testutils.Expect.True(t, strings.Contains(err.Error(), "already exists"), "Error should indicate tag already exists")
}

func TestCreateReleaseTag_TagNameFormat(t *testing.T) {
	tests := []struct {
		version     string
		expectedTag string
	}{
		{"1.0.0", "v1.0.0"},
		{"2.5.3", "v2.5.3"},
		{"0.1.0", "v0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			repo := testutils.SetupTestRepo(t)
			worktree, err := repo.Worktree()
			if err != nil {
				t.Fatalf("Failed to get worktree: %v", err)
			}

			repoPath := worktree.Filesystem.Root()
			version := &changelog.Version{
				Number: tt.version,
				Date:   "2024-01-15",
				Sections: []changelog.Section{
					{
						Type:    "added",
						Entries: []string{"Feature"},
					},
				},
			}

			err = createReleaseTag(repoPath, tt.version, version)
			if err != nil {
				t.Fatalf("createReleaseTag() error = %v", err)
			}

			_, err = repo.Tag(tt.expectedTag)
			if err != nil {
				t.Errorf("Tag %s should exist, got error: %v", tt.expectedTag, err)
			}
		})
	}
}

func TestBuildTagMessage(t *testing.T) {
	version := &changelog.Version{
		Number: "1.2.3",
		Date:   "2024-01-15",
		Sections: []changelog.Section{
			{
				Type: "added",
				Entries: []string{
					"Feature A",
					"Feature B",
				},
			},
			{
				Type:    "changed",
				Entries: []string{"Updated API"},
			},
			{
				Type: "fixed",
				Entries: []string{
					"Bug 1",
					"Bug 2",
				},
			},
		},
	}
	message := buildTagMessage("1.2.3", version)

	testutils.Expect.True(t, strings.HasPrefix(message, "Release 1.2.3\n\n"), "Message should start with release header")

	testutils.Expect.True(t, strings.Contains(message, "Added:\n"), "Should contain Added section")
	testutils.Expect.True(t, strings.Contains(message, "Changed:\n"), "Should contain Changed section")
	testutils.Expect.True(t, strings.Contains(message, "Fixed:\n"), "Should contain Fixed section")

	testutils.Expect.True(t, strings.Contains(message, "- Feature A\n"), "Should contain entry")
	testutils.Expect.True(t, strings.Contains(message, "- Feature B\n"), "Should contain entry")
	testutils.Expect.True(t, strings.Contains(message, "- Updated API\n"), "Should contain entry")
	testutils.Expect.True(t, strings.Contains(message, "- Bug 1\n"), "Should contain entry")
	testutils.Expect.True(t, strings.Contains(message, "- Bug 2\n"), "Should contain entry")

	sections := strings.Split(message, "\n\n")
	testutils.Expect.True(t, len(sections) >= 3, "Sections should be separated by blank lines")
}

func TestBuildTagMessage_EmptyVersion(t *testing.T) {
	version := &changelog.Version{
		Number:   "1.0.0",
		Date:     "2024-01-15",
		Sections: []changelog.Section{},
	}
	message := buildTagMessage("1.0.0", version)

	testutils.Expect.True(t, strings.HasPrefix(message, "Release 1.0.0\n\n"), "Should still have release header even with no sections")
}
