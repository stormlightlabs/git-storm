package main

import (
	"fmt"

	"github.com/stormlightlabs/git-storm/internal/toolchain"
)

func updateToolchainTargets(repoPath, newVersion string, selectors []string) ([]toolchain.Manifest, error) {
	if len(selectors) == 0 {
		return nil, nil
	}

	selected, interactive, available, err := toolchain.ResolveTargets(repoPath, selectors)
	if err != nil {
		return nil, err
	}

	if interactive {
		if len(available) == 0 {
			return nil, fmt.Errorf("no toolchain manifests detected for interactive selection")
		}
		chosen, err := toolchain.SelectManifests(available)
		if err != nil {
			return nil, err
		}
		selected = append(selected, chosen...)
	}

	selected = dedupeManifests(selected)
	if len(selected) == 0 {
		return nil, nil
	}

	for _, manifest := range selected {
		if err := toolchain.UpdateManifest(manifest, newVersion); err != nil {
			return nil, err
		}
	}

	return selected, nil
}

func dedupeManifests(manifests []toolchain.Manifest) []toolchain.Manifest {
	if len(manifests) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var result []toolchain.Manifest
	for _, manifest := range manifests {
		key := manifest.Path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, manifest)
	}
	return result
}
