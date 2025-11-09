package versioning

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stormlightlabs/git-storm/internal/changelog"
)

// BumpType represents the semantic version component to increment.
type BumpType string

const (
	BumpMajor BumpType = "major"
	BumpMinor BumpType = "minor"
	BumpPatch BumpType = "patch"
)

// Version represents a semantic version split into numeric components.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Parse converts a semantic version string (X.Y.Z) into a Version structure.
// An empty string returns 0.0.0 to simplify bump workflows.
func Parse(version string) (Version, error) {
	if version == "" {
		return Version{}, nil
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid semantic version: %s", version)
	}

	vals := make([]int, 3)
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return Version{}, fmt.Errorf("invalid semantic version: %s", version)
		}
		vals[i] = value
	}

	return Version{Major: vals[0], Minor: vals[1], Patch: vals[2]}, nil
}

// String formats the Version back into X.Y.Z form.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Bump increments the requested component following semver rules.
func (v Version) Bump(kind BumpType) Version {
	switch kind {
	case BumpMajor:
		return Version{Major: v.Major + 1, Minor: 0, Patch: 0}
	case BumpMinor:
		return Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
	case BumpPatch:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	default:
		return v
	}
}

// Next returns the bumped version string for the provided semantic version.
func Next(current string, kind BumpType) (string, error) {
	parsed, err := Parse(current)
	if err != nil {
		return "", err
	}
	return parsed.Bump(kind).String(), nil
}

// ParseBumpType validates user input into a BumpType.
func ParseBumpType(value string) (BumpType, error) {
	switch strings.ToLower(value) {
	case string(BumpMajor):
		return BumpMajor, nil
	case string(BumpMinor):
		return BumpMinor, nil
	case string(BumpPatch):
		return BumpPatch, nil
	default:
		return "", fmt.Errorf("invalid bump type %q (expected major, minor, or patch)", value)
	}
}

// LatestVersion scans a parsed changelog for the most recent released version.
// It skips the "Unreleased" section if present and validates with Keep a Changelog semantics.
func LatestVersion(ch *changelog.Changelog) (string, bool) {
	if ch == nil {
		return "", false
	}

	for _, v := range ch.Versions {
		if strings.EqualFold(v.Number, "unreleased") {
			continue
		}
		if err := changelog.ValidateVersion(v.Number); err == nil {
			return v.Number, true
		}
	}

	return "", false
}
