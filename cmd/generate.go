/*
USAGE

	storm generate [from] [to] [options]

FLAGS

	-i, --interactive       Review generated entries in a TUI
	    --since <tag>       Generate changes since the given tag
	-o, --output <path>     Write generated changelog to path
	    --repo <path>       Path to the Git repository (default: .)
*/
package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/style"
	"github.com/stormlightlabs/git-storm/internal/ui"
)

var (
	interactive bool
	sinceTag    string
)

// TODO(determinism): Add deduplication logic using diff-based identity
//
// Currently generates duplicate .changes/*.md files when:
// 1. Running generate multiple times on the same range
// 2. History is rewritten (rebase/amend) but commit content unchanged
//
// Implementation:
//
//  1. Before processing commits, load existing entries:
//     existingEntries := changeset.LoadExisting(".changes/data")
//     // Returns map[diffHash]Metadata for O(1) lookups
//
//  2. For each selected commit:
//     a. Compute diff hash: diffHash := changeset.ComputeDiffHash(repo, commit)
//     b. Check if exists: if meta, exists := existingEntries[diffHash]; exists {
//     - Same commit hash → true duplicate, skip
//     - Different commit hash → rebased/cherry-picked
//     * If --update-rebased: update metadata.CommitHash in JSON
//     * If --skip-rebased: skip (default)
//     * If --warn-rebased: print warning and skip
//     }
//     c. If not exists: create new entry with diff hash as filename
//
// 3. Report statistics:
//   - N new entries created
//   - M duplicates skipped (same commit)
//   - K rebased commits detected (same diff, different commit)
//
// Flags to add:
// --update-rebased    Update commit hash for rebased entries
// --skip-rebased      Skip rebased commits (default)
// --warn-rebased      Print warnings for rebased commits
// --force             Regenerate all entries (ignore existing)
//
// Related: See internal/changeset/changeset.go TODO for implementation details
func generateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "generate [from] [to]",
		Short: "Generate changelog entries from Git commits",
		Long: `Scans commits between two Git refs (tags or hashes) and outputs draft
entries in .changes/. Supports conventional commit parsing and
interactive review mode.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var from, to string

			if sinceTag != "" {
				from = sinceTag
				if len(args) > 0 {
					to = args[0]
				} else {
					to = "HEAD"
				}
			} else if len(args) == 0 {
				return fmt.Errorf("must specify either --since flag or [from] [to] arguments")
			} else {
				from, to = gitlog.ParseRefArgs(args)
			}

			repo, err := git.PlainOpen(repoPath)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			commits, err := gitlog.GetCommitRange(repo, from, to)
			if err != nil {
				return err
			}

			if len(commits) == 0 {
				style.Headlinef("No commits found between %s and %s", from, to)
				return nil
			}

			parser := &gitlog.ConventionalParser{}
			var selectedItems []ui.CommitItem

			if interactive {
				model := ui.NewCommitSelectorModel(commits, from, to, parser)
				p := tea.NewProgram(model, tea.WithAltScreen())

				finalModel, err := p.Run()
				if err != nil {
					return fmt.Errorf("failed to run interactive selector: %w", err)
				}

				selectorModel, ok := finalModel.(ui.CommitSelectorModel)
				if !ok {
					return fmt.Errorf("unexpected model type")
				}

				if selectorModel.IsCancelled() {
					style.Headline("Operation cancelled")
					return nil
				}

				selectedItems = selectorModel.GetSelectedItems()

				if len(selectedItems) == 0 {
					style.Headline("No commits selected")
					return nil
				}

				style.Headlinef("Generating entries for %d selected commits", len(selectedItems))
			} else {
				style.Headlinef("Found %d commits between %s and %s", len(commits), from, to)

				for _, commit := range commits {
					subject := commit.Message
					body := ""
					lines := strings.Split(commit.Message, "\n")
					if len(lines) > 0 {
						subject = lines[0]
						if len(lines) > 1 {
							body = strings.Join(lines[1:], "\n")
						}
					}

					meta, err := parser.Parse(commit.Hash.String(), subject, body, commit.Author.When)
					if err != nil {
						style.Println("Warning: failed to parse commit %s: %v", commit.Hash.String()[:gitlog.ShaLen], err)
						continue
					}

					category := parser.Categorize(meta)
					if category == "" {
						continue
					}

					selectedItems = append(selectedItems, ui.CommitItem{
						Commit:   commit,
						Meta:     meta,
						Category: category,
						Selected: true,
					})
				}
			}

			entries := []changeset.Entry{}
			skipped := 0

			for _, item := range selectedItems {
				if item.Category == "" {
					skipped++
					continue
				}

				entry := changeset.Entry{
					Type:     item.Category,
					Scope:    item.Meta.Scope,
					Summary:  item.Meta.Description,
					Breaking: item.Meta.Breaking,
				}

				entries = append(entries, entry)
			}

			changesDir := ".changes"
			created := 0
			for _, entry := range entries {
				filePath, err := changeset.Write(changesDir, entry)
				if err != nil {
					fmt.Printf("Error: failed to write entry: %v\n", err)
					continue
				}
				style.Addedf("✓ Created %s", filePath)
				created++
			}

			style.Newline()
			style.Headlinef("Generated %d changelog entries", created)
			if skipped > 0 {
				style.Println("Skipped %d commits (reverts or non-matching types)", skipped)
			}

			return nil
		},
	}

	c.Flags().BoolVarP(&interactive, "interactive", "i", false, "Review changes interactively in a TUI")
	c.Flags().StringVar(&sinceTag, "since", "", "Generate changes since the given tag")

	return c
}
