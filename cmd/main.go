package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/style"
)

var (
	repoPath string
	output   string
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

	if err := fang.Execute(ctx, root, fang.WithColorSchemeFunc(style.NewColorScheme)); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
