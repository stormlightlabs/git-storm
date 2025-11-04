package main

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/diff"
	"github.com/stormlightlabs/git-storm/internal/ui"
)

var (
	repoPath string
	output   string
)

var (
	fromRef     string
	toRef       string
	interactive bool
	sinceTag    string
)

var (
	changeType string
	scope      string
	summary    string
	outputJSON bool
)

var (
	releaseVersion string
	tagRelease     bool
	dryRun         bool
)

const versionString string = "0.1.0-dev"

// parseRefArgs parses command arguments to extract from/to refs.
// Supports both "from..to" and "from to" syntax.
func parseRefArgs(args []string) (from, to string) {
	if len(args) == 1 {
		parts := strings.Split(args[0], "..")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return args[0], "HEAD"
	}
	return args[0], args[1]
}

// runDiff executes the diff command by reading file contents from two git refs and launching the TUI.
func runDiff(fromRef, toRef, filePath string, expanded bool) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var filesToDiff []string
	if filePath != "" {
		filesToDiff = []string{filePath}
	} else {
		filesToDiff, err = getChangedFiles(repo, fromRef, toRef)
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
		oldContent, err := getFileContent(repo, fromRef, file)
		if err != nil {
			oldContent = ""
		}

		newContent, err := getFileContent(repo, toRef, file)
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

	model := ui.NewMultiFileDiffModel(allDiffs, expanded)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI failed: %w", err)
	}

	return nil
}

// getFileContent reads the content of a file at a specific ref (commit, tag, or branch).
func getFileContent(repo *git.Repository, ref, filePath string) (string, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", ref, err)
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	file, err := tree.File(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	content, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// getChangedFiles returns the list of files that changed between two commits.
func getChangedFiles(repo *git.Repository, fromRef, toRef string) ([]string, error) {
	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", fromRef, err)
	}

	toHash, err := repo.ResolveRevision(plumbing.Revision(toRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", toRef, err)
	}

	fromCommit, err := repo.CommitObject(*fromHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", fromRef, err)
	}

	toCommit, err := repo.CommitObject(*toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", toRef, err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for %s: %w", fromRef, err)
	}

	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for %s: %w", toRef, err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	files := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.To.Name != "" {
			files = append(files, change.To.Name)
		} else {
			files = append(files, change.From.Name)
		}
	}

	return files, nil
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current storm version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(versionString)
			return nil
		},
	}
}

func diffCmd() *cobra.Command {
	var filePath string
	var expanded bool

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
			from, to := parseRefArgs(args)
			return runDiff(from, to, filePath, expanded)
		},
	}

	c.Flags().StringVarP(&filePath, "file", "f", "", "Specific file to diff (optional, shows all files if omitted)")
	c.Flags().BoolVarP(&expanded, "expanded", "e", false, "Show all unchanged lines (disable compression)")

	return c
}

func releaseCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "release",
		Short: "Promote unreleased changes into a new changelog version",
		Long: `Merges all .changes entries into CHANGELOG.md under a new version header.
Optionally creates a Git tag and clears the .changes directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("release command not implemented")
			fmt.Printf("version=%v tag=%v dry-run=%v\n", releaseVersion, tagRelease, dryRun)
			return nil
		},
	}

	c.Flags().StringVar(&releaseVersion, "version", "", "Semantic version for the new release (e.g., 1.3.0)")
	c.Flags().BoolVar(&tagRelease, "tag", false, "Create a Git tag after release")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	c.MarkFlagRequired("version")

	return c
}

func unreleasedCmd() *cobra.Command {
	add := &cobra.Command{
		Use:   "add",
		Short: "Add a new unreleased change entry",
		Long: `Creates a new .changes/<date>-<summary>.md file with the specified type,
scope, and summary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("unreleased add not implemented")
			fmt.Printf("type=%q scope=%q summary=%q\n", changeType, scope, summary)
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
			fmt.Println("unreleased list not implemented")
			fmt.Printf("outputJSON=%v\n", outputJSON)
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
			fmt.Println("unreleased review not implemented (TUI)")
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

func generateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "generate [from] [to]",
		Short: "Generate changelog entries from Git commits",
		Long: `Scans commits between two Git refs (tags or hashes) and outputs draft
entries in .changes/. Supports conventional commit parsing and
interactive review mode.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("generate command not implemented")
			fmt.Printf("from=%v to=%v interactive=%v sinceTag=%v\n", fromRef, toRef, interactive, sinceTag)
			return nil
		},
	}

	c.Flags().BoolVarP(&interactive, "interactive", "i", false, "Review changes interactively in a TUI")
	c.Flags().StringVar(&sinceTag, "since", "", "Generate changes since the given tag")

	return c
}

func main() {
	ctx := context.Background()
	root := &cobra.Command{
		Use:   "storm",
		Short: "A Git-aware changelog manager for Go projects",
		Long: `storm is a modern changelog generator inspired by Towncrier.
It manages .changes/ entries, generates Keep a Changelog sections,
and can review commits interactively through a TUI.`,
	}

	root.PersistentFlags().StringVar(&repoPath, "repo", ".", "Path to the Git repository")
	root.PersistentFlags().StringVarP(&output, "output", "o", "CHANGELOG.md", "Output changelog file path")
	root.AddCommand(generateCmd(), unreleasedCmd(), releaseCmd(), diffCmd(), versionCmd())

	if err := fang.Execute(ctx, root); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
