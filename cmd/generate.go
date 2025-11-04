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

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changeset"
	"github.com/stormlightlabs/git-storm/internal/gitlog"
	"github.com/stormlightlabs/git-storm/internal/style"
)

var (
	interactive bool
	sinceTag    string
)

// getCommitRange returns commits reachable from toRef but not from fromRef.
// This implements the git log from..to range semantics.
func getCommitRange(repo *git.Repository, fromRef, toRef string) ([]*object.Commit, error) {
	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", fromRef, err)
	}

	toHash, err := repo.ResolveRevision(plumbing.Revision(toRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: %w", toRef, err)
	}

	toCommits := make(map[plumbing.Hash]bool)
	toIter, err := repo.Log(&git.LogOptions{From: *toHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", toRef, err)
	}

	err = toIter.ForEach(func(c *object.Commit) error {
		toCommits[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits from %s: %w", toRef, err)
	}

	fromCommits := make(map[plumbing.Hash]bool)
	fromIter, err := repo.Log(&git.LogOptions{From: *fromHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", fromRef, err)
	}

	err = fromIter.ForEach(func(c *object.Commit) error {
		fromCommits[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits from %s: %w", fromRef, err)
	}

	// Collect commits that are in toCommits but not in fromCommits
	result := []*object.Commit{}
	toIter, err = repo.Log(&git.LogOptions{From: *toHash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits from %s: %w", toRef, err)
	}

	err = toIter.ForEach(func(c *object.Commit) error {
		if !fromCommits[c.Hash] {
			result = append(result, c)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to collect commit range: %w", err)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
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
			} else if len(args) == 1 {
				parts := strings.Split(args[0], "..")
				if len(parts) == 2 {
					from, to = parts[0], parts[1]
				} else {
					from, to = args[0], "HEAD"
				}
			} else {
				from, to = args[0], args[1]
			}

			if interactive {
				style.Headline("Interactive mode not yet implemented")
				fmt.Println("Will generate entries in non-interactive mode...")
			}

			repo, err := git.PlainOpen(repoPath)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			commits, err := getCommitRange(repo, from, to)
			if err != nil {
				return err
			}

			if len(commits) == 0 {
				style.Headline(fmt.Sprintf("No commits found between %s and %s", from, to))
				return nil
			}

			style.Headline(fmt.Sprintf("Found %d commits between %s and %s", len(commits), from, to))

			parser := &gitlog.ConventionalParser{}
			entries := []changeset.Entry{}
			skipped := 0

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
					fmt.Printf("Warning: failed to parse commit %s: %v\n", commit.Hash.String()[:7], err)
					continue
				}

				category := parser.Categorize(meta)
				if category == "" {
					skipped++
					continue
				}

				entry := changeset.Entry{
					Type:     category,
					Scope:    meta.Scope,
					Summary:  meta.Description,
					Breaking: meta.Breaking,
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

			fmt.Println()
			style.Headline(fmt.Sprintf("Generated %d changelog entries", created))
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
