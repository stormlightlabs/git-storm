package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stormlightlabs/git-storm/internal/changeset"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		wantVersionCount int
		wantFirstVersion string
		wantFirstDate    string
	}{
		{
			name: "empty file returns default header",
			content: `# Changelog

All notable changes to this project will be documented in this file.`,
			wantVersionCount: 0,
		},
		{
			name: "single version with sections",
			content: `# Changelog

## [1.0.0] - 2025-01-15

### Added
- New feature A
- New feature B

### Fixed
- Bug fix C
`,
			wantVersionCount: 1,
			wantFirstVersion: "1.0.0",
			wantFirstDate:    "2025-01-15",
		},
		{
			name: "multiple versions",
			content: `# Changelog

## [Unreleased]

## [1.2.0] - 2025-01-15

### Added
- Feature X

## [1.1.0] - 2025-01-10

### Fixed
- Bug Y
`,
			wantVersionCount: 3,
			wantFirstVersion: "Unreleased",
			wantFirstDate:    "Unreleased",
		},
		{
			name: "version with comparison links",
			content: `# Changelog

## [1.0.0] - 2025-01-15

### Added
- Feature A

[1.0.0]: https://github.com/user/repo/releases/tag/v1.0.0
`,
			wantVersionCount: 1,
			wantFirstVersion: "1.0.0",
			wantFirstDate:    "2025-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

			if err := os.WriteFile(changelogPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			changelog, err := Parse(changelogPath)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(changelog.Versions) != tt.wantVersionCount {
				t.Errorf("Version count = %d, want %d", len(changelog.Versions), tt.wantVersionCount)
			}

			if tt.wantVersionCount > 0 {
				if changelog.Versions[0].Number != tt.wantFirstVersion {
					t.Errorf("First version = %s, want %s", changelog.Versions[0].Number, tt.wantFirstVersion)
				}
				if changelog.Versions[0].Date != tt.wantFirstDate {
					t.Errorf("First date = %s, want %s", changelog.Versions[0].Date, tt.wantFirstDate)
				}
			}
		})
	}
}

func TestParseNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "NONEXISTENT.md")

	changelog, err := Parse(changelogPath)
	if err != nil {
		t.Fatalf("Parse() should not error on non-existent file: %v", err)
	}

	if len(changelog.Versions) != 0 {
		t.Errorf("Empty changelog should have 0 versions, got %d", len(changelog.Versions))
	}

	if !strings.Contains(changelog.Header, "Keep a Changelog") {
		t.Errorf("Default header should contain 'Keep a Changelog'")
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name           string
		entries        []changeset.Entry
		version        string
		date           string
		wantSectionCnt int
		wantFirstType  string
		wantBreaking   bool
	}{
		{
			name: "single entry",
			entries: []changeset.Entry{
				{Type: "added", Summary: "New feature"},
			},
			version:        "1.0.0",
			date:           "2025-01-15",
			wantSectionCnt: 1,
			wantFirstType:  "added",
			wantBreaking:   false,
		},
		{
			name: "multiple types in correct order",
			entries: []changeset.Entry{
				{Type: "fixed", Summary: "Bug fix"},
				{Type: "added", Summary: "New feature"},
				{Type: "changed", Summary: "Updated API"},
			},
			version:        "2.0.0",
			date:           "2025-01-20",
			wantSectionCnt: 3,
			wantFirstType:  "added",
		},
		{
			name: "entry with scope",
			entries: []changeset.Entry{
				{Type: "added", Scope: "cli", Summary: "New command"},
			},
			version:        "1.1.0",
			date:           "2025-01-18",
			wantSectionCnt: 1,
			wantFirstType:  "added",
		},
		{
			name: "breaking change",
			entries: []changeset.Entry{
				{Type: "changed", Summary: "API change", Breaking: true},
			},
			version:        "2.0.0",
			date:           "2025-02-01",
			wantSectionCnt: 1,
			wantFirstType:  "changed",
			wantBreaking:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := Build(tt.entries, tt.version, tt.date)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			if version.Number != tt.version {
				t.Errorf("Version number = %s, want %s", version.Number, tt.version)
			}

			if version.Date != tt.date {
				t.Errorf("Version date = %s, want %s", version.Date, tt.date)
			}

			if len(version.Sections) != tt.wantSectionCnt {
				t.Errorf("Section count = %d, want %d", len(version.Sections), tt.wantSectionCnt)
			}

			if tt.wantSectionCnt > 0 {
				if version.Sections[0].Type != tt.wantFirstType {
					t.Errorf("First section type = %s, want %s", version.Sections[0].Type, tt.wantFirstType)
				}

				if tt.wantBreaking {
					firstEntry := version.Sections[0].Entries[0]
					if !strings.Contains(firstEntry, "**BREAKING:**") {
						t.Errorf("Breaking change should have **BREAKING:** prefix, got: %s", firstEntry)
					}
				}
			}
		})
	}
}

func TestBuildInvalidVersion(t *testing.T) {
	entries := []changeset.Entry{{Type: "added", Summary: "Test"}}

	invalidVersions := []string{
		"v1.0.0",
		"1.0",
		"1.0.0.0",
		"abc",
	}

	for _, version := range invalidVersions {
		t.Run("invalid_version_"+version, func(t *testing.T) {
			_, err := Build(entries, version, "2025-01-15")
			if err == nil {
				t.Errorf("Build() should error for invalid version %s", version)
			}
		})
	}
}

func TestBuildInvalidDate(t *testing.T) {
	entries := []changeset.Entry{{Type: "added", Summary: "Test"}}

	invalidDates := []string{
		"2025-13-01",
		"2025-01-32",
		"01-15-2025",
		"2025/01/15",
		"not-a-date",
	}

	for _, date := range invalidDates {
		t.Run("invalid_date_"+date, func(t *testing.T) {
			_, err := Build(entries, "1.0.0", date)
			if err == nil {
				t.Errorf("Build() should error for invalid date %s", date)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name              string
		existingVersions  []Version
		newVersion        Version
		wantPositionIndex int
	}{
		{
			name:              "merge into empty changelog",
			existingVersions:  []Version{},
			newVersion:        Version{Number: "1.0.0", Date: "2025-01-15"},
			wantPositionIndex: 0,
		},
		{
			name: "merge below unreleased",
			existingVersions: []Version{
				{Number: "Unreleased", Date: "Unreleased"},
				{Number: "1.0.0", Date: "2025-01-10"},
			},
			newVersion:        Version{Number: "1.1.0", Date: "2025-01-15"},
			wantPositionIndex: 1,
		},
		{
			name: "merge at top when no unreleased",
			existingVersions: []Version{
				{Number: "1.0.0", Date: "2025-01-10"},
			},
			newVersion:        Version{Number: "1.1.0", Date: "2025-01-15"},
			wantPositionIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changelog := &Changelog{
				Versions: tt.existingVersions,
			}

			Merge(changelog, &tt.newVersion)

			if changelog.Versions[tt.wantPositionIndex].Number != tt.newVersion.Number {
				t.Errorf("Version at position %d = %s, want %s",
					tt.wantPositionIndex,
					changelog.Versions[tt.wantPositionIndex].Number,
					tt.newVersion.Number)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	changelog := &Changelog{
		Header: "# Changelog\n\nTest changelog",
		Versions: []Version{
			{
				Number: "1.0.0",
				Date:   "2025-01-15",
				Sections: []Section{
					{
						Type:    "added",
						Entries: []string{"New feature A", "New feature B"},
					},
					{
						Type:    "fixed",
						Entries: []string{"Bug fix C"},
					},
				},
			},
		},
	}

	err := Write(changelogPath, changelog, tmpDir)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		t.Fatalf("CHANGELOG.md was not created")
	}

	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("Failed to read CHANGELOG.md: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "# Changelog") {
		t.Errorf("Missing header")
	}
	if !strings.Contains(contentStr, "## [1.0.0] - 2025-01-15") {
		t.Errorf("Missing version header")
	}
	if !strings.Contains(contentStr, "### Added") {
		t.Errorf("Missing Added section")
	}
	if !strings.Contains(contentStr, "### Fixed") {
		t.Errorf("Missing Fixed section")
	}
	if !strings.Contains(contentStr, "- New feature A") {
		t.Errorf("Missing entry: New feature A")
	}
	if !strings.Contains(contentStr, "- Bug fix C") {
		t.Errorf("Missing entry: Bug fix C")
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"1.0.0", false},
		{"0.1.0", false},
		{"10.20.30", false},
		{"v1.0.0", true},
		{"1.0", true},
		{"1.0.0.0", true},
		{"1.x.0", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%s) error = %v, wantErr %v", tt.version, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		date    string
		wantErr bool
	}{
		{"2025-01-15", false},
		{"2024-12-31", false},
		{"2025-13-01", true},
		{"2025-01-32", true},
		{"01-15-2025", true},
		{"2025/01/15", true},
		{"not-a-date", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			err := ValidateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate(%s) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      string
	}{
		{
			name:      "https format",
			remoteURL: "https://github.com/user/repo.git",
			want:      "https://github.com/user/repo",
		},
		{
			name:      "https without .git",
			remoteURL: "https://github.com/user/repo",
			want:      "https://github.com/user/repo",
		},
		{
			name:      "ssh format",
			remoteURL: "git@github.com:user/repo.git",
			want:      "https://github.com/user/repo",
		},
		{
			name:      "ssh without .git",
			remoteURL: "git@github.com:user/repo",
			want:      "https://github.com/user/repo",
		},
		{
			name:      "non-github url",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGitHubURL(tt.remoteURL)
			if got != tt.want {
				t.Errorf("parseGitHubURL(%s) = %s, want %s", tt.remoteURL, got, tt.want)
			}
		})
	}
}

func TestSectionOrdering(t *testing.T) {
	entries := []changeset.Entry{
		{Type: "security", Summary: "Security fix"},
		{Type: "removed", Summary: "Removed feature"},
		{Type: "fixed", Summary: "Bug fix"},
		{Type: "changed", Summary: "Changed behavior"},
		{Type: "added", Summary: "New feature"},
	}

	version, err := Build(entries, "1.0.0", "2025-01-15")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expectedOrder := []string{"added", "changed", "removed", "fixed", "security"}
	if len(version.Sections) != len(expectedOrder) {
		t.Fatalf("Expected %d sections, got %d", len(expectedOrder), len(version.Sections))
	}

	for i, expectedType := range expectedOrder {
		if version.Sections[i].Type != expectedType {
			t.Errorf("Section %d: got type %s, want %s", i, version.Sections[i].Type, expectedType)
		}
	}
}

func TestEntrySorting(t *testing.T) {
	entries := []changeset.Entry{
		{Type: "added", Summary: "Zebra feature"},
		{Type: "added", Summary: "Apple feature"},
		{Type: "added", Summary: "Mango feature"},
	}

	version, err := Build(entries, "1.0.0", "2025-01-15")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(version.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(version.Sections))
	}

	sortedEntries := version.Sections[0].Entries
	if len(sortedEntries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(sortedEntries))
	}

	if !strings.Contains(sortedEntries[0], "Apple") {
		t.Errorf("First entry should contain 'Apple', got: %s", sortedEntries[0])
	}
	if !strings.Contains(sortedEntries[1], "Mango") {
		t.Errorf("Second entry should contain 'Mango', got: %s", sortedEntries[1])
	}
	if !strings.Contains(sortedEntries[2], "Zebra") {
		t.Errorf("Third entry should contain 'Zebra', got: %s", sortedEntries[2])
	}
}

func TestScopeFormatting(t *testing.T) {
	entries := []changeset.Entry{
		{Type: "added", Scope: "cli", Summary: "New command"},
	}

	version, err := Build(entries, "1.0.0", "2025-01-15")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	entry := version.Sections[0].Entries[0]
	if !strings.Contains(entry, "**cli:**") {
		t.Errorf("Entry should contain formatted scope, got: %s", entry)
	}
}
