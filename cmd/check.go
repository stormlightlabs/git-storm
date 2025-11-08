/*
USAGE

	storm check [from] [to] [options]

FLAGS

	--since <tag>       Check changes since the given tag
	--repo <path>       Path to the Git repository (default: .)

# DESCRIPTION

Validates that all commits in the specified range have corresponding unreleased
changelog entries. This is useful for CI enforcement to ensure developers
document their changes.

Commits containing [nochanges] or [skip changelog] in the message are skipped.

Exit codes:

	0 - All commits have changelog entries
	1 - One or more commits are missing changelog entries
	2 - Command execution error

TODO(issue-linking): Support checking for issue numbers in entries when --issue flag is implemented in `unreleased partial`.

  - This requires integrating with at Gitea/Forgejo, Github, Gitlab, and Tangled
*/
package main

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/style"
)

// checkCmd validates that all commits in a range have corresponding changelog entries.
func checkCmd() *cobra.Command {
	var sinceTag string

	c := &cobra.Command{
		Use:   "check [from] [to]",
		Short: "Validate changelog entries exist for all commits",
		Long: `Checks that all commits in the specified range have corresponding
.changes/*.md entries. Useful for CI enforcement.

Commits with [nochanges] or [skip changelog] in their message are skipped.`,
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

			changesDir := ".changes"
			existingMetadata, err := changeset.LoadExistingMetadata(changesDir)
			if err != nil {
				return fmt.Errorf("failed to load existing metadata: %w", err)
			}

			style.Headlinef("Checking %d commits between %s and %s", len(commits), from, to)
			style.Newline()

			var missingEntries []string
			skippedCount := 0

			for _, commit := range commits {
				message := strings.ToLower(commit.Message)
				if strings.Contains(message, "[nochanges]") || strings.Contains(message, "[skip changelog]") {
					skippedCount++
					continue
				}

				diffHash, err := changeset.ComputeDiffHash(commit)
				if err != nil {
					style.Println("Warning: failed to compute diff hash for commit %s: %v", commit.Hash.String()[:7], err)
					continue
				}

				if _, exists := existingMetadata[diffHash]; !exists {
					sha7 := commit.Hash.String()[:7]
					subject := strings.Split(commit.Message, "\n")[0]
					missingEntries = append(missingEntries, fmt.Sprintf("%s - %s", sha7, subject))
				}
			}

			if len(missingEntries) == 0 {
				style.Addedf("✓ All commits have changelog entries")
				if skippedCount > 0 {
					style.Println("  Skipped %d commits with [nochanges] marker", skippedCount)
				}
				return nil
			}

			style.Println("%s", style.StyleRemoved.Render(fmt.Sprintf("✗ %d commits missing changelog entries:", len(missingEntries))))
			style.Newline()

			for _, entry := range missingEntries {
				style.Println("  - %s", entry)
			}

			style.Newline()
			style.Println("To create entries, run:")
			style.Println("  storm generate %s %s --interactive", from, to)
			style.Println("Or manually create entries with:")
			style.Println("  storm unreleased partial <commit-ref>")

			return fmt.Errorf("changelog validation failed")
		},
	}

	c.Flags().StringVar(&sinceTag, "since", "", "Check changes since the given tag")
	return c
}
