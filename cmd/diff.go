/*
NAME

	storm diff — Display an inline diff between two refs or commits.

SYNOPSIS

	storm diff <from>..<to> [options]
	storm diff <from> <to>   [options]

DESCRIPTION

	Displays an inline diff highlighting added, removed, and unchanged lines
	between two refs or commits.

	Supports multiple input formats:
	  • Range syntax:         commit1..commit2
	  • Separate arguments:   commit1 commit2
	  • Truncated hashes:     7de6f6d..18363c2

	If --file is not specified, storm shows all changed files with pagination.

	By default, large blocks of unchanged lines are compressed. Use --expanded
	to show all lines, or toggle this interactively with ‘e’ in the TUI.
*/
package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/diff"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/tty"
	"github.com/stormlightlabs/git-storm/internal/ui"
)

func diffCmd() *cobra.Command {
	var filePath string
	var expanded bool
	var viewName string

	c := &cobra.Command{
		Use:   "diff <from>..<to> | diff <from> <to>",
		Short: "Show a line-based diff between two commits or tags",
		Long: `Displays an inline diff (added/removed/unchanged lines) between two refs.

Supports multiple input formats:
  - Range syntax: commit1..commit2
  - Separate args: commit1 commit2
  - Truncated hashes: 7de6f6d..18363c2

If --file is not specified, shows all changed files with pagination.

By default, large blocks of unchanged lines are compressed. Use --expanded
to show all lines. You can also toggle this with 'e' in the TUI.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			from, to := gitlog.ParseRefArgs(args)
			viewKind, err := parseDiffView(viewName)
			if err != nil {
				return err
			}
			return runDiff(from, to, filePath, expanded, viewKind)
		},
	}

	c.Flags().StringVarP(&filePath, "file", "f", "", "Specific file to diff (optional, shows all files if omitted)")
	c.Flags().BoolVarP(&expanded, "expanded", "e", false, "Show all unchanged lines (disable compression)")
	c.Flags().StringVarP(&viewName, "view", "v", "split", "Diff rendering: split or unified")

	return c
}

// runDiff executes the diff command by reading file contents from two git refs and launching the TUI.
func runDiff(fromRef, toRef, filePath string, expanded bool, view diff.DiffViewKind) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var filesToDiff []string
	if filePath != "" {
		filesToDiff = []string{filePath}
	} else {
		filesToDiff, err = gitlog.GetChangedFiles(repo, fromRef, toRef)
		if err != nil {
			return fmt.Errorf("failed to get changed files: %w", err)
		}
		if len(filesToDiff) == 0 {
			fmt.Println("No files changed between", fromRef, "and", toRef)
			return nil
		}
	}

	allDiffs := make([]ui.FileDiff, 0, len(filesToDiff))

	for _, file := range filesToDiff {
		oldContent, err := gitlog.GetFileContent(repo, fromRef, file)
		if err != nil {
			oldContent = ""
		}

		newContent, err := gitlog.GetFileContent(repo, toRef, file)
		if err != nil {
			newContent = ""
		}

		oldLines := strings.Split(oldContent, "\n")
		newLines := strings.Split(newContent, "\n")

		myers := &diff.Myers{}
		edits, err := myers.Compute(oldLines, newLines)
		if err != nil {
			return fmt.Errorf("diff computation failed for %s: %w", file, err)
		}

		allDiffs = append(allDiffs, ui.FileDiff{
			Edits:   edits,
			OldPath: fromRef + ":" + file,
			NewPath: toRef + ":" + file,
		})
	}

	if !tty.IsInteractive() {
		return outputPlainDiff(allDiffs, expanded, view)
	}

	model := ui.NewMultiFileDiffModel(allDiffs, expanded, view)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI failed: %w", err)
	}

	return nil
}

func parseDiffView(viewName string) (diff.DiffViewKind, error) {
	switch strings.ToLower(strings.TrimSpace(viewName)) {
	case "", "split", "side-by-side", "s":
		return diff.ViewSplit, nil
	case "unified", "u":
		return diff.ViewUnified, nil
	default:
		return 0, fmt.Errorf("invalid view %q: expected one of split, unified", viewName)
	}
}

// outputPlainDiff outputs diffs in plain text format for non-interactive environments.
//
// TODO: move this to package [diff]
func outputPlainDiff(allDiffs []ui.FileDiff, expanded bool, view diff.DiffViewKind) error {
	for i, fileDiff := range allDiffs {
		fmt.Printf("=== File %d/%d ===\n", i+1, len(allDiffs))
		fmt.Printf("--- %s\n", fileDiff.OldPath)
		fmt.Printf("+++ %s\n", fileDiff.NewPath)
		fmt.Println()

		var formatter diff.Formatter
		switch view {
		case diff.ViewUnified:
			formatter = &diff.UnifiedFormatter{
				TerminalWidth:   80,
				ShowLineNumbers: true,
				Expanded:        expanded,
				EnableWordWrap:  false,
			}
		default:
			formatter = &diff.SideBySideFormatter{
				TerminalWidth:   80,
				ShowLineNumbers: true,
				Expanded:        expanded,
				EnableWordWrap:  false,
			}
		}

		output := formatter.Format(fileDiff.Edits)
		fmt.Println(output)

		if i < len(allDiffs)-1 {
			fmt.Println()
		}
	}

	return nil
}
