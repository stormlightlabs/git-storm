/*
USAGE

	storm release --version <X.Y.Z> [options]

FLAGS

	--version <X.Y.Z>     Semantic version for the new release (required)
	--date <YYYY-MM-DD>   Release date (default: today)
	--clear-changes       Delete .changes/*.md files after successful release
	--dry-run             Preview changes without writing files
	--tag                 Create a Git tag after release (not implemented)
	--repo <path>         Path to the Git repository (default: .)
	--output <path>       Output changelog file path (default: CHANGELOG.md)
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changelog"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/style"
)

func releaseCmd() *cobra.Command {
	var (
		version      string
		date         string
		clearChanges bool
		dryRun       bool
		tag          bool
	)

	c := &cobra.Command{
		Use:   "release",
		Short: "Promote unreleased changes into a new changelog version",
		Long: `Merges all .changes entries into CHANGELOG.md under a new version header.
Optionally creates a Git tag and clears the .changes directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := changelog.ValidateVersion(version); err != nil {
				return err
			}

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

			changelogPath := filepath.Join(repoPath, output)
			existingChangelog, err := changelog.Parse(changelogPath)
			if err != nil {
				return fmt.Errorf("failed to parse existing changelog: %w", err)
			}

			changelog.Merge(existingChangelog, newVersion)

			if dryRun {
				style.Headline("Dry-run mode: Preview of CHANGELOG.md")
				style.Newline()
				displayVersionPreview(newVersion)
				style.Newline()
				style.Println("No files were modified (--dry-run)")
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

			style.Newline()
			style.Headlinef("Release %s completed successfully", version)

			if tag {
				style.Newline()
				style.Println("Note: --tag flag is not yet implemented (Phase 7)")
			}

			return nil
		},
	}

	c.Flags().StringVar(&version, "version", "", "Semantic version for the new release (e.g., 1.3.0)")
	c.Flags().StringVar(&date, "date", "", "Release date in YYYY-MM-DD format (default: today)")
	c.Flags().BoolVar(&clearChanges, "clear-changes", false, "Delete .changes/*.md files after successful release")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	c.Flags().BoolVar(&tag, "tag", false, "Create a Git tag after release (not implemented)")
	c.MarkFlagRequired("version")

	return c
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
