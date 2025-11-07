// Package changelog implements Keep a Changelog parsing, building, and writing.
//
// It generates CHANGELOG.md files compliant with https://keepachangelog.com/en/1.1.0/
package changelog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/stormlightlabs/git-storm/internal/changeset"
)

// Changelog represents the entire CHANGELOG.md file structure.
type Changelog struct {
	Header   string    // Preamble text before versions
	Versions []Version // All versions in chronological order (newest first)
	Links    []string  // Version comparison links at the bottom
}

// Version represents a single version section in the changelog.
type Version struct {
	Number   string    // Semantic version (e.g., "1.2.0")
	Date     string    // ISO date (YYYY-MM-DD) or "Unreleased"
	Sections []Section // Category sections (Added, Changed, etc.)
}

// Section represents a category section within a version.
type Section struct {
	Type    string   // added, changed, deprecated, removed, fixed, security
	Entries []string // Individual entries without leading dashes
}

// sectionOrder defines the Keep a Changelog section ordering.
var sectionOrder = []string{"added", "changed", "deprecated", "removed", "fixed", "security"}

// sectionTitles maps internal types to Keep a Changelog titles.
var sectionTitles = map[string]string{
	"added":      "Added",
	"changed":    "Changed",
	"deprecated": "Deprecated",
	"removed":    "Removed",
	"fixed":      "Fixed",
	"security":   "Security",
}

// versionHeaderRegex matches version headers like "## [1.2.0] - 2025-01-15" or "## [Unreleased]"
var versionHeaderRegex = regexp.MustCompile(`^##\s+\[([^\]]+)\](?:\s+-\s+(.+))?$`)

// sectionHeaderRegex matches section headers like "### Added"
var sectionHeaderRegex = regexp.MustCompile(`^###\s+(.+)$`)

// entryRegex matches changelog entries like "- Entry text"
var entryRegex = regexp.MustCompile(`^-\s+(.+)$`)

// semanticVersionRegex validates semantic versioning (X.Y.Z)
var semanticVersionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// linkRegex matches comparison links like "[1.2.0]: https://..."
var linkRegex = regexp.MustCompile(`^\[([^\]]+)\]:\s+(.+)$`)

// Parse reads and parses an existing CHANGELOG.md file.
// Returns an empty Changelog with default header if the file doesn't exist.
func Parse(path string) (*Changelog, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return newEmptyChangelog(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open changelog: %w", err)
	}
	defer file.Close()

	changelog := &Changelog{}
	scanner := bufio.NewScanner(file)

	var headerLines []string
	var currentVersion *Version
	var currentSection *Section
	inLinks := false

	for scanner.Scan() {
		line := scanner.Text()

		if linkMatch := linkRegex.FindStringSubmatch(line); linkMatch != nil {
			inLinks = true
			changelog.Links = append(changelog.Links, line)
			continue
		}

		if inLinks {
			if strings.TrimSpace(line) != "" {
				changelog.Links = append(changelog.Links, line)
			}
			continue
		}

		if versionMatch := versionHeaderRegex.FindStringSubmatch(line); versionMatch != nil {
			if currentVersion != nil {
				if currentSection != nil && len(currentSection.Entries) > 0 {
					currentVersion.Sections = append(currentVersion.Sections, *currentSection)
				}
				changelog.Versions = append(changelog.Versions, *currentVersion)
			}

			currentVersion = &Version{
				Number: versionMatch[1],
			}
			if len(versionMatch) > 2 && versionMatch[2] != "" {
				currentVersion.Date = versionMatch[2]
			} else {
				currentVersion.Date = "Unreleased"
			}
			currentSection = nil
			continue
		}

		if sectionMatch := sectionHeaderRegex.FindStringSubmatch(line); sectionMatch != nil {
			if currentVersion != nil {
				if currentSection != nil && len(currentSection.Entries) > 0 {
					currentVersion.Sections = append(currentVersion.Sections, *currentSection)
				}

				sectionTitle := sectionMatch[1]
				sectionType := findSectionType(sectionTitle)
				currentSection = &Section{
					Type:    sectionType,
					Entries: []string{},
				}
			}
			continue
		}

		if entryMatch := entryRegex.FindStringSubmatch(line); entryMatch != nil {
			if currentSection != nil {
				currentSection.Entries = append(currentSection.Entries, entryMatch[1])
			}
			continue
		}

		if currentVersion == nil {
			headerLines = append(headerLines, line)
		}
	}

	if currentVersion != nil {
		if currentSection != nil && len(currentSection.Entries) > 0 {
			currentVersion.Sections = append(currentVersion.Sections, *currentSection)
		}
		changelog.Versions = append(changelog.Versions, *currentVersion)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read changelog: %w", err)
	}

	changelog.Header = strings.TrimSpace(strings.Join(headerLines, "\n"))
	if changelog.Header == "" {
		changelog.Header = defaultHeader()
	}

	return changelog, nil
}

// Build creates a new Version from changeset entries.
//
// Entries are grouped by type, sorted, and formatted with breaking change prefixes.
func Build(entries []changeset.Entry, version, date string) (*Version, error) {
	if err := ValidateVersion(version); err != nil {
		return nil, err
	}

	if err := ValidateDate(date); err != nil {
		return nil, err
	}

	grouped := make(map[string][]string)
	for _, entry := range entries {
		text := entry.Summary
		if entry.Scope != "" {
			text = fmt.Sprintf("**%s:** %s", entry.Scope, text)
		}
		if entry.Breaking {
			text = fmt.Sprintf("**BREAKING:** %s", text)
		}

		grouped[entry.Type] = append(grouped[entry.Type], text)
	}

	for typ := range grouped {
		sort.Strings(grouped[typ])
	}

	// Build sections in Keep a Changelog order
	var sections []Section
	for _, typ := range sectionOrder {
		if entryList, exists := grouped[typ]; exists && len(entryList) > 0 {
			sections = append(sections, Section{
				Type:    typ,
				Entries: entryList,
			})
		}
	}

	return &Version{
		Number:   version,
		Date:     date,
		Sections: sections,
	}, nil
}

// Merge inserts a new version into the changelog at the top (below Unreleased if present).
func Merge(changelog *Changelog, version *Version) {
	insertIndex := 0
	if len(changelog.Versions) > 0 && strings.ToLower(changelog.Versions[0].Number) == "unreleased" {
		insertIndex = 1
	}

	versions := make([]Version, 0, len(changelog.Versions)+1)
	versions = append(versions, changelog.Versions[:insertIndex]...)
	versions = append(versions, *version)
	versions = append(versions, changelog.Versions[insertIndex:]...)
	changelog.Versions = versions
}

// Write writes the changelog to a file with proper Keep a Changelog formatting.
//
// Generates version comparison links if a git remote is available.
func Write(path string, changelog *Changelog, repoPath string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create changelog: %w", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	if changelog.Header != "" {
		fmt.Fprintf(w, "%s\n\n", changelog.Header)
	}

	for i, version := range changelog.Versions {
		if i > 0 {
			fmt.Fprintln(w)
		}

		if version.Date == "" || strings.ToLower(version.Date) == "unreleased" {
			fmt.Fprintf(w, "## [%s]\n\n", version.Number)
		} else {
			fmt.Fprintf(w, "## [%s] - %s\n\n", version.Number, version.Date)
		}

		for j, section := range version.Sections {
			if j > 0 {
				fmt.Fprintln(w)
			}

			title := sectionTitles[section.Type]
			if title == "" {
				if len(section.Type) > 0 {
					title = strings.ToUpper(section.Type[:1]) + section.Type[1:]
				} else {
					title = section.Type
				}
			}
			fmt.Fprintf(w, "### %s\n\n", title)

			for _, entry := range section.Entries {
				fmt.Fprintf(w, "- %s\n", entry)
			}
		}
	}

	links, err := GenerateLinks(repoPath, changelog.Versions)
	if err == nil && len(links) > 0 {
		fmt.Fprintln(w)
		for _, link := range links {
			fmt.Fprintln(w, link)
		}
	} else if len(changelog.Links) > 0 {
		fmt.Fprintln(w)
		for _, link := range changelog.Links {
			fmt.Fprintln(w, link)
		}
	}

	return nil
}

// GenerateLinks creates version comparison links for GitHub repositories.
func GenerateLinks(repoPath string, versions []Version) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, fmt.Errorf("no origin remote configured: %w", err)
	}

	if len(remote.Config().URLs) == 0 {
		return nil, fmt.Errorf("no remote URL configured")
	}

	remoteURL := remote.Config().URLs[0]
	baseURL := parseGitHubURL(remoteURL)
	if baseURL == "" {
		return nil, fmt.Errorf("not a GitHub repository")
	}

	var links []string
	for i, version := range versions {
		var link string
		if strings.ToLower(version.Number) == "unreleased" {
			if len(versions) > 1 {
				link = fmt.Sprintf("[Unreleased]: %s/compare/v%s...HEAD", baseURL, versions[1].Number)
			} else {
				link = fmt.Sprintf("[Unreleased]: %s/compare/HEAD", baseURL)
			}
		} else {
			if i+1 < len(versions) && strings.ToLower(versions[i+1].Number) != "unreleased" {
				link = fmt.Sprintf("[%s]: %s/compare/v%s...v%s", version.Number, baseURL, versions[i+1].Number, version.Number)
			} else {
				link = fmt.Sprintf("[%s]: %s/releases/tag/v%s", version.Number, baseURL, version.Number)
			}
		}
		links = append(links, link)
	}

	return links, nil
}

// ValidateVersion checks if a version string follows semantic versioning (X.Y.Z).
func ValidateVersion(version string) error {
	if !semanticVersionRegex.MatchString(version) {
		return fmt.Errorf("invalid semantic version '%s': must be X.Y.Z format (e.g., 1.2.0)", version)
	}
	return nil
}

// ValidateDate checks if a date string follows ISO 8601 format (YYYY-MM-DD).
func ValidateDate(date string) error {
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date '%s': must be YYYY-MM-DD format", date)
	}
	return nil
}

// parseGitHubURL extracts the base GitHub URL from a git remote URL.
//
// Handles both HTTPS and SSH formats.
func parseGitHubURL(remoteURL string) string {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	if strings.HasPrefix(remoteURL, "https://github.com/") {
		return remoteURL
	}

	if parts, ok := strings.CutPrefix(remoteURL, "git@github.com:"); ok {
		return "https://github.com/" + parts
	}
	return ""
}

// findSectionType converts a section title to its internal type.
func findSectionType(title string) string {
	titleLower := strings.ToLower(strings.TrimSpace(title))
	for typ, standardTitle := range sectionTitles {
		if strings.ToLower(standardTitle) == titleLower {
			return typ
		}
	}
	return titleLower
}

// newEmptyChangelog creates a changelog with default header and empty versions.
func newEmptyChangelog() *Changelog {
	return &Changelog{
		Header:   defaultHeader(),
		Versions: []Version{},
		Links:    []string{},
	}
}

// defaultHeader returns the standard Keep a Changelog header.
func defaultHeader() string {
	return `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).`
}
