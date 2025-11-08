/*
USAGE

	storm unreleased <subcommand> [options]

SUBCOMMANDS

	add         Add a new unreleased change entry
	list        List all unreleased changes
	review      Review unreleased changes interactively
	partial     Create entry linked to a specific commit

USAGE

	storm unreleased add [options]

FLAGS

	--type <type>       Change type (added, changed, fixed, removed, security)
	--scope <scope>     Optional subsystem or module name
	--summary <text>    Short description of the change
	--repo <path>       Path to the repository (default: .)

USAGE

	storm unreleased list [options]

FLAGS

	--json              Output as JSON
	--repo <path>       Path to the repository (default: .)

USAGE

	storm unreleased review [options]

FLAGS

	--repo <path>       Path to the repository (default: .)
	--output <file>     Optional file to export reviewed notes

USAGE

	storm unreleased partial <commit-ref> [options]

FLAGS

	--type <type>       Override change type (auto-detected from commit message)
	--summary <text>    Override summary (auto-detected from commit message)
	--scope <scope>     Optional subsystem or module name
	--repo <path>       Path to the repository (default: .)
*/
package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/style"
	"github.com/stormlightlabs/git-storm/internal/ui"
)

func unreleasedCmd() *cobra.Command {
	var (
		changeType string
		scope      string
		summary    string
		outputJSON bool
	)

	changesDir := ".changes"
	validTypes := []string{"added", "changed", "fixed", "removed", "security"}

	add := &cobra.Command{
		Use:   "add",
		Short: "Add a new unreleased change entry",
		Long: `Creates a new .changes/<date>-<summary>.md file with the specified type,
scope, and summary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains(validTypes, changeType) {
				return fmt.Errorf("invalid type %q: must be one of %s", changeType, strings.Join(validTypes, ", "))
			}

			if filePath, err := changeset.Write(changesDir, changeset.Entry{
				Type:    changeType,
				Scope:   scope,
				Summary: summary,
			}); err != nil {
				return fmt.Errorf("failed to create changelog entry: %w", err)
			} else {
				style.Addedf("Created %s", filePath)
				return nil
			}
		},
	}
	add.Flags().StringVar(&changeType, "type", "", "Type of change (added, changed, fixed, removed, security)")
	add.Flags().StringVar(&scope, "scope", "", "Optional scope or subsystem name")
	add.Flags().StringVar(&summary, "summary", "", "Short summary of the change")
	add.MarkFlagRequired("type")
	add.MarkFlagRequired("summary")

	list := &cobra.Command{
		Use:   "list",
		Short: "List all unreleased changes",
		Long:  "Prints all pending .changes entries to stdout. Supports JSON output.",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := changeset.List(changesDir)
			if err != nil {
				return fmt.Errorf("failed to list changelog entries: %w", err)
			}

			if len(entries) == 0 {
				style.Println("No unreleased changes found")
				return nil
			}

			if outputJSON {
				jsonBytes, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal entries to JSON: %w", err)
				}
				fmt.Println(string(jsonBytes))
				return nil
			}

			style.Headlinef("Found %d unreleased change(s):", len(entries))
			style.Newline()

			for _, e := range entries {
				displayEntry(e)
			}

			return nil
		},
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "Output results as JSON")

	review := &cobra.Command{
		Use:   "review",
		Short: "Review unreleased changes interactively",
		Long: `Launches an interactive Bubble Tea TUI to review, edit, or categorize
unreleased entries before final release.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := changeset.List(changesDir)
			if err != nil {
				return fmt.Errorf("failed to list changelog entries: %w", err)
			}

			if len(entries) == 0 {
				style.Println("No unreleased changes found")
				return nil
			}

			model := ui.NewChangesetReviewModel(entries)
			p := tea.NewProgram(model, tea.WithAltScreen())

			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("failed to run review TUI: %w", err)
			}

			reviewModel, ok := finalModel.(ui.ChangesetReviewModel)
			if !ok {
				return fmt.Errorf("unexpected model type")
			}

			if reviewModel.IsCancelled() {
				style.Headline("Review cancelled")
				return nil
			}

			items := reviewModel.GetReviewedItems()
			deleteCount := 0
			editCount := 0

			for _, item := range items {
				if item.Action == ui.ActionDelete {
					if err := changeset.Delete(changesDir, item.Entry.Filename); err != nil {
						return fmt.Errorf("failed to delete %s: %w", item.Entry.Filename, err)
					}
					deleteCount++
					style.Successf("Deleted: %s", item.Entry.Filename)
				}
			}

			for _, item := range items {
				if item.Action == ui.ActionEdit {
					editorModel := ui.NewEntryEditorModel(item.Entry)
					p := tea.NewProgram(editorModel, tea.WithAltScreen())

					finalModel, err := p.Run()
					if err != nil {
						return fmt.Errorf("failed to run editor TUI: %w", err)
					}

					editor, ok := finalModel.(ui.EntryEditorModel)
					if !ok {
						return fmt.Errorf("unexpected model type")
					}

					if editor.IsCancelled() {
						style.Warningf("Skipped editing: %s", item.Entry.Filename)
						continue
					}

					if editor.IsConfirmed() {
						editedEntry := editor.GetEditedEntry()
						if err := changeset.Update(changesDir, item.Entry.Filename, editedEntry); err != nil {
							return fmt.Errorf("failed to update %s: %w", item.Entry.Filename, err)
						}
						editCount++
						style.Successf("Updated: %s", item.Entry.Filename)
					}
				}
			}

			if deleteCount == 0 && editCount == 0 {
				style.Headline("No changes requested")
				return nil
			}

			style.Headlinef("Review completed: %d deleted, %d edited", deleteCount, editCount)
			return nil
		},
	}

	partial := &cobra.Command{
		Use:   "partial <commit-ref>",
		Short: "Create entry linked to a specific commit",
		Long: `Creates a new .changes/<sha7>.<type>.md file based on the specified commit.
Auto-detects type and summary from conventional commit format, with optional overrides.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commitRef := args[0]

			repo, err := git.PlainOpen(repoPath)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			hash, err := repo.ResolveRevision(plumbing.Revision(commitRef))
			if err != nil {
				return fmt.Errorf("failed to resolve commit ref %q: %w", commitRef, err)
			}

			commit, err := repo.CommitObject(*hash)
			if err != nil {
				return fmt.Errorf("failed to get commit object: %w", err)
			}

			parser := &gitlog.ConventionalParser{}
			subject := commit.Message
			body := ""
			lines := strings.Split(commit.Message, "\n")
			if len(lines) > 0 {
				subject = lines[0]
				if len(lines) > 1 {
					body = strings.Join(lines[1:], "\n")
				}
			}

			meta, err := parser.Parse(hash.String(), subject, body, commit.Author.When)
			if err != nil {
				return fmt.Errorf("failed to parse commit message: %w", err)
			}

			category := parser.Categorize(meta)

			if changeType != "" {
				if !slices.Contains(validTypes, changeType) {
					return fmt.Errorf("invalid type %q: must be one of %s", changeType, strings.Join(validTypes, ", "))
				}
				category = changeType
			} else if category == "" {
				return fmt.Errorf("could not auto-detect change type from commit message, please specify --type")
			}

			entrySummary := meta.Description
			if summary != "" {
				entrySummary = summary
			}

			if scope != "" {
				meta.Scope = scope
			}

			sha7 := hash.String()[:7]
			filename := fmt.Sprintf("%s.%s.md", sha7, category)
			filePath := changesDir + "/" + filename

			entry := changeset.Entry{
				Type:       category,
				Scope:      meta.Scope,
				Summary:    entrySummary,
				Breaking:   meta.Breaking,
				CommitHash: hash.String(),
			}

			if _, err := changeset.WritePartial(changesDir, filename, entry); err != nil {
				return fmt.Errorf("failed to create changelog entry: %w", err)
			}

			style.Addedf("Created %s", filePath)
			return nil
		},
	}
	partial.Flags().StringVar(&changeType, "type", "", "Override change type (auto-detected from commit)")
	partial.Flags().StringVar(&scope, "scope", "", "Optional scope or subsystem name")
	partial.Flags().StringVar(&summary, "summary", "", "Override summary (auto-detected from commit)")

	root := &cobra.Command{
		Use:   "unreleased",
		Short: "Manage unreleased changes (.changes directory)",
		Long: `Work with unreleased change notes. Supports adding, listing,
and reviewing pending entries before release.`,
	}
	root.AddCommand(add, list, review, partial)
	return root
}

// displayEntry formats and prints a single changelog entry with color-coded type.
func displayEntry(e changeset.EntryWithFile) {
	var typeLabel string
	switch e.Entry.Type {
	case "added":
		typeLabel = style.StyleAdded.Render(fmt.Sprintf("[%s]", e.Entry.Type))
	case "changed":
		typeLabel = style.StyleChanged.Render(fmt.Sprintf("[%s]", e.Entry.Type))
	case "fixed":
		typeLabel = style.StyleFixed.Render(fmt.Sprintf("[%s]", e.Entry.Type))
	case "removed":
		typeLabel = style.StyleRemoved.Render(fmt.Sprintf("[%s]", e.Entry.Type))
	case "security":
		typeLabel = style.StyleSecurity.Render(fmt.Sprintf("[%s]", e.Entry.Type))
	default:
		typeLabel = fmt.Sprintf("[%s]", e.Entry.Type)
	}

	var scopePart string
	if e.Entry.Scope != "" {
		scopePart = fmt.Sprintf("(%s) ", e.Entry.Scope)
	}

	style.Println("%s %s%s", typeLabel, scopePart, e.Entry.Summary)
	style.Println("  File: %s", e.Filename)
	if e.Entry.Breaking {
		style.Println("  Breaking: %s\n", style.StyleRemoved.Render("YES"))
	}
	style.Newline()
}
