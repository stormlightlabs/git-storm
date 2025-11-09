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

// TODO: use ldflags
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
	root.AddCommand(generateCmd(), unreleasedCmd(), releaseCmd(), bumpCmd(), diffCmd(), checkCmd(), versionCmd())

	if err := fang.Execute(ctx, root, fang.WithColorSchemeFunc(style.NewColorScheme)); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
