package versioning

import (
	"testing"

	"github.com/stormlightlabs/git-storm/internal/changelog"
)

func TestParseAndBump(t *testing.T) {
	version, err := Parse("1.2.3")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if version.Major != 1 || version.Minor != 2 || version.Patch != 3 {
		t.Fatalf("unexpected parse result: %#v", version)
	}

	major := version.Bump(BumpMajor)
	if got := major.String(); got != "2.0.0" {
		t.Fatalf("major bump = %s, want 2.0.0", got)
	}

	minor := version.Bump(BumpMinor)
	if got := minor.String(); got != "1.3.0" {
		t.Fatalf("minor bump = %s, want 1.3.0", got)
	}

	patch := version.Bump(BumpPatch)
	if got := patch.String(); got != "1.2.4" {
		t.Fatalf("patch bump = %s, want 1.2.4", got)
	}
}

func TestParseBumpType(t *testing.T) {
	cases := map[string]BumpType{
		"MAJOR": BumpMajor,
		"minor": BumpMinor,
		"Patch": BumpPatch,
	}

	for input, expected := range cases {
		kind, err := ParseBumpType(input)
		if err != nil {
			t.Fatalf("ParseBumpType(%s) returned error: %v", input, err)
		}
		if kind != expected {
			t.Fatalf("ParseBumpType(%s) = %s, want %s", input, kind, expected)
		}
	}

	if _, err := ParseBumpType("invalid"); err == nil {
		t.Fatal("expected error for invalid bump type")
	}
}

func TestNextHandlesEmptyVersion(t *testing.T) {
	got, err := Next("", BumpMinor)
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if got != "0.1.0" {
		t.Fatalf("Next = %s, want 0.1.0", got)
	}
}

func TestLatestVersion(t *testing.T) {
	ch := &changelog.Changelog{
		Versions: []changelog.Version{
			{Number: "Unreleased"},
			{Number: "1.5.0"},
			{Number: "0.9.1"},
		},
	}

	version, ok := LatestVersion(ch)
	if !ok {
		t.Fatal("LatestVersion returned false")
	}
	if version != "1.5.0" {
		t.Fatalf("LatestVersion = %s, want 1.5.0", version)
	}

	if _, ok := LatestVersion(&changelog.Changelog{}); ok {
		t.Fatal("expected LatestVersion to return false when no releases exist")
	}
}
