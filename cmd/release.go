/*
USAGE

	storm release --version <X.Y.Z> [options]

FLAGS

	--version <X.Y.Z>     Semantic version for the new release (required)
	--bump <type>         Automatically bump the previous version (major|minor|patch)
	--date <YYYY-MM-DD>   Release date (default: today)
	--clear-changes       Delete .changes/*.md files after successful release
	--dry-run             Preview changes without writing files
	--tag                 Create an annotated Git tag with release notes
	--toolchain <value>   Update toolchain manifests (path/type or 'interactive')
	--repo <path>         Path to the Git repository (default: .)
	--output <path>       Output changelog file path (default: CHANGELOG.md)
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changelog"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/shared"
	"github.com/stormlightlabs/git-storm/internal/style"
	"github.com/stormlightlabs/git-storm/internal/versioning"
)

func releaseCmd() *cobra.Command {
	var (
		version      string
		bumpKind     string
		date         string
		clearChanges bool
		dryRun       bool
		tag          bool
		toolchains   []string
	)

	c := &cobra.Command{
		Use:   "release",
		Short: "Promote unreleased changes into a new changelog version",
		Long: `Merges all .changes entries into CHANGELOG.md under a new version header.
Optionally creates a Git tag and clears the .changes directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			changelogPath := filepath.Join(repoPath, output)
			existingChangelog, err := changelog.Parse(changelogPath)
			if err != nil {
				return fmt.Errorf("failed to parse changelog: %w", err)
			}

			resolvedVersion, err := resolveReleaseVersion(version, bumpKind, existingChangelog)
			if err != nil {
				return err
			}
			version = resolvedVersion

			releaseDate := date
			if releaseDate == "" {
				releaseDate = time.Now().Format("2006-01-02")
			} else {
				if err := changelog.ValidateDate(releaseDate); err != nil {
					return err
				}
			}

			style.Headlinef("Preparing release %s (%s)", version, releaseDate)
			style.Newline()

			changesDir := ".changes"
			entries, err := changeset.List(changesDir)
			if err != nil {
				return fmt.Errorf("failed to read .changes directory: %w", err)
			}

			if len(entries) == 0 {
				return fmt.Errorf("no unreleased changes found in %s", changesDir)
			}

			style.Println("Found %d unreleased entries", len(entries))
			style.Newline()

			var entryList []changeset.Entry
			for _, e := range entries {
				entryList = append(entryList, e.Entry)
			}

			newVersion, err := changelog.Build(entryList, version, releaseDate)
			if err != nil {
				return fmt.Errorf("failed to build version: %w", err)
			}

			changelog.Merge(existingChangelog, newVersion)

			if dryRun {
				style.Headline("Dry-run mode: Preview of CHANGELOG.md")
				style.Newline()
				displayVersionPreview(newVersion)
				style.Newline()
				style.Println("No files were modified (--dry-run)")
				if len(toolchains) > 0 {
					style.Warningf("Skipping toolchain updates (--dry-run)")
				}
				return nil
			}

			if err := changelog.Write(changelogPath, existingChangelog, repoPath); err != nil {
				return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
			}

			style.Addedf("✓ Updated %s", changelogPath)

			if clearChanges {
				deletedCount := 0
				for _, entry := range entries {
					filePath := filepath.Join(changesDir, entry.Filename)
					if err := os.Remove(filePath); err != nil {
						style.Println("Warning: failed to delete %s: %v", filePath, err)
						continue
					}
					deletedCount++
				}
				style.Println("✓ Deleted %d entry files from %s", deletedCount, changesDir)
			}

			if len(toolchains) > 0 {
				updated, err := updateToolchainTargets(repoPath, version, toolchains)
				if err != nil {
					return err
				}
				for _, manifest := range updated {
					style.Addedf("✓ Updated %s", manifest.RelPath)
				}
			}

			style.Newline()
			style.Headlinef("Release %s completed successfully", version)

			if tag {
				style.Newline()
				if err := createReleaseTag(repoPath, version, newVersion); err != nil {
					return fmt.Errorf("failed to create Git tag: %w", err)
				}
				tagName := fmt.Sprintf("v%s", version)
				style.Addedf("✓ Created Git tag %s", tagName)
			}

			return nil
		},
	}

	c.Flags().StringVar(&version, "version", "", "Semantic version for the new release (e.g., 1.3.0)")
	c.Flags().StringVar(&bumpKind, "bump", "", "Automatically bump the previous version (major, minor, or patch)")
	c.Flags().StringVar(&date, "date", "", "Release date in YYYY-MM-DD format (default: today)")
	c.Flags().BoolVar(&clearChanges, "clear-changes", false, "Delete .changes/*.md files after successful release")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	c.Flags().BoolVar(&tag, "tag", false, "Create an annotated Git tag with release notes")
	c.Flags().StringSliceVar(&toolchains, "toolchain", nil, "Toolchain manifests to update (paths, types, or 'interactive')")

	return c
}

func resolveReleaseVersion(versionFlag, bumpFlag string, existing *changelog.Changelog) (string, error) {
	if bumpFlag == "" {
		if versionFlag == "" {
			return "", fmt.Errorf("either --version or --bump must be provided")
		}
		if err := changelog.ValidateVersion(versionFlag); err != nil {
			return "", err
		}
		return versionFlag, nil
	}

	if versionFlag != "" {
		return "", fmt.Errorf("--version and --bump cannot be used together")
	}

	kind, err := versioning.ParseBumpType(bumpFlag)
	if err != nil {
		return "", err
	}

	var current string
	if v, ok := versioning.LatestVersion(existing); ok {
		current = v
	}

	return versioning.Next(current, kind)
}

// createReleaseTag creates an annotated Git tag for the release with changelog entries as the message.
func createReleaseTag(repoPath, version string, versionData *changelog.Version) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	tagName := fmt.Sprintf("v%s", version)

	_, err = repo.Tag(tagName)
	if err == nil {
		return fmt.Errorf("tag %s already exists", tagName)
	}

	tagMessage := buildTagMessage(version, versionData)

	_, err = repo.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: tagMessage,
		Tagger: &object.Signature{
			Name:  "storm",
			Email: "noreply@storm",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// buildTagMessage formats the version's changelog entries into a tag message.
func buildTagMessage(version string, versionData *changelog.Version) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Release %s\n\n", version))

	for i, section := range versionData.Sections {
		if i > 0 {
			builder.WriteString("\n")
		}

		sectionTitle := shared.TitleCase(section.Type)
		builder.WriteString(fmt.Sprintf("%s:\n", sectionTitle))

		for _, entry := range section.Entries {
			builder.WriteString(fmt.Sprintf("- %s\n", entry))
		}
	}

	return builder.String()
}

// displayVersionPreview shows a formatted preview of the version being released.
func displayVersionPreview(version *changelog.Version) {
	fmt.Printf("## [%s] - %s\n\n", version.Number, version.Date)

	for i, section := range version.Sections {
		if i > 0 {
			fmt.Println()
		}

		var sectionTitle string
		switch section.Type {
		case "added":
			sectionTitle = style.StyleAdded.Render("### Added")
		case "changed":
			sectionTitle = style.StyleChanged.Render("### Changed")
		case "deprecated":
			sectionTitle = "### Deprecated"
		case "removed":
			sectionTitle = style.StyleRemoved.Render("### Removed")
		case "fixed":
			sectionTitle = style.StyleFixed.Render("### Fixed")
		case "security":
			sectionTitle = style.StyleSecurity.Render("### Security")
		default:
			sectionTitle = fmt.Sprintf("### %s", section.Type)
		}
		fmt.Println(sectionTitle)
		fmt.Println()

		for _, entry := range section.Entries {
			fmt.Printf("- %s\n", entry)
		}
	}
}
