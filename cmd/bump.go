package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/git-storm/internal/changelog"
	"github.com/stormlightlabs/git-storm/internal/style"
	"github.com/stormlightlabs/git-storm/internal/versioning"
)

func bumpCmd() *cobra.Command {
	var bumpKind string
	var toolchainSelectors []string

	cmd := &cobra.Command{
		Use:   "bump",
		Short: "Calculate the next semantic version and optionally update toolchain manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := versioning.ParseBumpType(bumpKind)
			if err != nil {
				return err
			}

			changelogPath := filepath.Join(repoPath, output)
			parsed, err := changelog.Parse(changelogPath)
			if err != nil {
				return fmt.Errorf("failed to parse changelog: %w", err)
			}

			current, _ := versioning.LatestVersion(parsed)
			nextVersion, err := versioning.Next(current, kind)
			if err != nil {
				return err
			}

			style.Headlinef("Next version: %s", nextVersion)

			updated, err := updateToolchainTargets(repoPath, nextVersion, toolchainSelectors)
			if err != nil {
				return err
			}
			for _, manifest := range updated {
				style.Addedf("✓ Updated %s", manifest.RelPath)
			}

			fmt.Fprintln(cmd.OutOrStdout(), nextVersion)
			return nil
		},
	}

	cmd.Flags().StringVar(&bumpKind, "bump", "", "Which semver component to bump (major, minor, or patch)")
	cmd.Flags().StringSliceVar(&toolchainSelectors, "toolchain", nil, "Toolchain manifests to update (paths, types, or 'interactive')")
	cmd.MarkFlagRequired("bump")

	return cmd
}
