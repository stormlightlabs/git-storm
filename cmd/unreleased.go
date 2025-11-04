/*
USAGE

	storm unreleased <subcommand> [options]

SUBCOMMANDS

	add         Add a new unreleased change entry
	list        List all unreleased changes
	review      Review unreleased changes interactively

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
*/
package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/style"
	"github.com/stormlightlabs/git-storm/internal/ui"
)

func unreleasedCmd() *cobra.Command {
	add := &cobra.Command{
		Use:   "add",
		Short: "Add a new unreleased change entry",
		Long: `Creates a new .changes/<date>-<summary>.md file with the specified type,
scope, and summary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			validTypes := []string{"added", "changed", "fixed", "removed", "security"}
			if !slices.Contains(validTypes, changeType) {
				return fmt.Errorf("invalid type %q: must be one of %s", changeType, strings.Join(validTypes, ", "))
			}

			entry := changeset.Entry{
				Type:    changeType,
				Scope:   scope,
				Summary: summary,
			}

			changesDir := ".changes"
			filePath, err := changeset.Write(changesDir, entry)
			if err != nil {
				return fmt.Errorf("failed to create changelog entry: %w", err)
			}

			style.Addedf("Created %s", filePath)
			return nil
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
			changesDir := ".changes"
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
			changesDir := ".changes"
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
				switch item.Action {
				case ui.ActionDelete:
					deleteCount++
				case ui.ActionEdit:
					editCount++
				}
			}

			if deleteCount == 0 && editCount == 0 {
				style.Headline("No changes requested")
				return nil
			}

			style.Headlinef("Review completed: %d to delete, %d to edit", deleteCount, editCount)
			style.Println("Note: Delete and edit actions are not yet implemented")

			return nil
		},
	}

	root := &cobra.Command{
		Use:   "unreleased",
		Short: "Manage unreleased changes (.changes directory)",
		Long: `Work with unreleased change notes. Supports adding, listing,
and reviewing pending entries before release.`,
	}
	root.AddCommand(add, list, review)

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
