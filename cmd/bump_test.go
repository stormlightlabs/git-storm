package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleChangelog = `# Changelog

## [Unreleased]

## [1.2.3] - 2024-01-01
### Added
- Initial release
`

func TestBumpCommandPrintsVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CHANGELOG.md"), sampleChangelog)

	oldRepo := repoPath
	oldOutput := output
	repoPath = dir
	output = "CHANGELOG.md"
	t.Cleanup(func() {
		repoPath = oldRepo
		output = oldOutput
	})

	cmd := bumpCmd()
	cmd.SetArgs([]string{"--bump", "minor"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bump command failed: %v", err)
	}

	if !strings.Contains(out.String(), "1.3.0") {
		t.Fatalf("expected output to contain 1.3.0, got %q", out.String())
	}
}

func TestBumpCommandUpdatesPackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CHANGELOG.md"), sampleChangelog)
	writeFile(t, filepath.Join(dir, "package.json"), `{"name":"demo","version":"1.2.3"}`)

	oldRepo := repoPath
	oldOutput := output
	repoPath = dir
	output = "CHANGELOG.md"
	t.Cleanup(func() {
		repoPath = oldRepo
		output = oldOutput
	})

	cmd := bumpCmd()
	cmd.SetArgs([]string{"--bump", "patch", "--toolchain", "package.json"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bump command failed: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}

	if !strings.Contains(string(contents), "1.2.4") {
		t.Fatalf("expected package.json to contain bumped version, got %s", contents)
	}
}

func writeFile(t *testing.T, path, contents string) {
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
