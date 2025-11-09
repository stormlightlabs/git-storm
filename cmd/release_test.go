package main

import (
	"encoding/json"
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

func TestResolveReleaseVersion(t *testing.T) {
	existing := &changelog.Changelog{Versions: []changelog.Version{{Number: "Unreleased"}, {Number: "1.2.3"}}}

	version, err := resolveReleaseVersion("", "minor", existing)
	if err != nil {
		t.Fatalf("resolveReleaseVersion returned error: %v", err)
	}
	if version != "1.3.0" {
		t.Fatalf("expected 1.3.0, got %s", version)
	}

	version, err = resolveReleaseVersion("2.0.0", "", existing)
	if err != nil {
		t.Fatalf("resolveReleaseVersion returned error: %v", err)
	}
	if version != "2.0.0" {
		t.Fatalf("expected 2.0.0, got %s", version)
	}

	if _, err := resolveReleaseVersion("2.0.0", "patch", existing); err == nil {
		t.Fatal("expected error when both --version and --bump are set")
	}

	blankChangelog := &changelog.Changelog{}
	version, err = resolveReleaseVersion("", "patch", blankChangelog)
	if err != nil {
		t.Fatalf("resolveReleaseVersion returned error: %v", err)
	}
	if version != "0.0.1" {
		t.Fatalf("expected 0.0.1 for empty changelog, got %s", version)
	}
}

func TestReleaseOutput_JSONStructure(t *testing.T) {
	output := ReleaseOutput{
		Version:        "1.0.0",
		Date:           "2024-01-15",
		EntriesCount:   3,
		ChangelogPath:  "CHANGELOG.md",
		TagCreated:     true,
		TagName:        "v1.0.0",
		ChangesCleared: true,
		DeletedCount:   3,
		DryRun:         false,
		VersionData: &changelog.Version{
			Number: "1.0.0",
			Date:   "2024-01-15",
			Sections: []changelog.Section{
				{
					Type:    "added",
					Entries: []string{"Feature 1"},
				},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	var unmarshaled ReleaseOutput
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	testutils.Expect.Equal(t, unmarshaled.Version, "1.0.0")
	testutils.Expect.Equal(t, unmarshaled.Date, "2024-01-15")
	testutils.Expect.Equal(t, unmarshaled.EntriesCount, 3)
	testutils.Expect.Equal(t, unmarshaled.TagCreated, true)
	testutils.Expect.Equal(t, unmarshaled.TagName, "v1.0.0")
	testutils.Expect.Equal(t, unmarshaled.ChangesCleared, true)
	testutils.Expect.Equal(t, unmarshaled.DeletedCount, 3)
}

func TestReleaseOutput_DryRunJSON(t *testing.T) {
	output := ReleaseOutput{
		Version:       "1.0.0",
		Date:          "2024-01-15",
		EntriesCount:  2,
		ChangelogPath: "CHANGELOG.md",
		DryRun:        true,
		VersionData: &changelog.Version{
			Number: "1.0.0",
			Date:   "2024-01-15",
			Sections: []changelog.Section{
				{
					Type:    "fixed",
					Entries: []string{"Bug fix"},
				},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	testutils.Expect.True(t, strings.Contains(string(jsonBytes), `"dry_run": true`))
	testutils.Expect.True(t, strings.Contains(string(jsonBytes), `"tag_created": false`))
	testutils.Expect.True(t, strings.Contains(string(jsonBytes), `"changes_cleared": false`))
}
