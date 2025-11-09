package toolchain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "Cargo.toml", `[package]
name = "demo"
version = "0.1.0"

[dependencies]
serde = "1"
`)

	subDir := filepath.Join(dir, "app")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	writeFile(t, subDir, "pyproject.toml", `[project]
name = "demo-app"
version = "0.2.0"
`)

	writeFile(t, dir, "package.json", `{
  "name": "demo-web",
  "version": "1.5.0"
}`)

	writeFile(t, dir, "deno.json", `{
  "version": "3.1.4"
}`)

	manifests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(manifests) != 4 {
		t.Fatalf("expected 4 manifests, got %d", len(manifests))
	}

	manifestByType := make(map[ManifestType]Manifest)
	for _, manifest := range manifests {
		manifestByType[manifest.Type] = manifest
	}

	if manifestByType[ManifestCargo].Version != "0.1.0" {
		t.Fatalf("cargo version mismatch: %#v", manifestByType[ManifestCargo])
	}
	if manifestByType[ManifestPython].Version != "0.2.0" {
		t.Fatalf("python version mismatch: %#v", manifestByType[ManifestPython])
	}
	if manifestByType[ManifestNode].Version != "1.5.0" {
		t.Fatalf("npm version mismatch: %#v", manifestByType[ManifestNode])
	}
	if manifestByType[ManifestDeno].Version != "3.1.4" {
		t.Fatalf("deno version mismatch: %#v", manifestByType[ManifestDeno])
	}
}

func TestResolveTargets(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[package]
name = "demo"
version = "0.1.0"
`)
	writeFile(t, dir, "package.json", `{"name":"demo","version":"1.0.0"}`)

	selected, interactive, available, err := ResolveTargets(dir, []string{"cargo"})
	if err != nil {
		t.Fatalf("ResolveTargets returned error: %v", err)
	}
	if interactive {
		t.Fatal("expected interactive to be false")
	}
	if len(selected) != 1 || selected[0].Type != ManifestCargo {
		t.Fatalf("expected cargo manifest to be selected, got %#v", selected)
	}
	if len(available) != 2 {
		t.Fatalf("expected 2 discovered manifests, got %d", len(available))
	}

	selected, interactive, _, err = ResolveTargets(dir, []string{"interactive"})
	if err != nil {
		t.Fatalf("ResolveTargets interactive: %v", err)
	}
	if !interactive {
		t.Fatal("expected interactive mode to be true")
	}
	if len(selected) != 0 {
		t.Fatalf("interactive request should not preselect manifests: got %d", len(selected))
	}

	selected, _, _, err = ResolveTargets(dir, []string{"package.json"})
	if err != nil {
		t.Fatalf("ResolveTargets path selection: %v", err)
	}
	if len(selected) != 1 || selected[0].Type != ManifestNode {
		t.Fatalf("expected package.json selection, got %#v", selected)
	}
}

func TestUpdateManifest(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[package]
name = "demo"
version = "0.1.0"
`)
	writeFile(t, dir, "pyproject.toml", `[project]
name = "demo"
version = "0.2.0"
`)
	writeFile(t, dir, "package.json", `{"name":"demo","version":"1.0.0"}`)
	writeFile(t, dir, "deno.json", `{"version":"1.1.0"}`)

	files := []string{"Cargo.toml", "pyproject.toml", "package.json", "deno.json"}
	for _, name := range files {
		manifest, err := loadManifest(dir, filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("loadManifest(%s) error: %v", name, err)
		}
		if err := UpdateManifest(manifest, "9.9.9"); err != nil {
			t.Fatalf("UpdateManifest(%s) error: %v", name, err)
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read updated %s: %v", name, err)
		}
		if !strings.Contains(string(data), "9.9.9") {
			t.Fatalf("%s was not updated: %s", name, string(data))
		}
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}
