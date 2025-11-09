package toolchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ManifestType enumerates supported ecosystem manifests whose versions we can bump.
type ManifestType string

const (
	ManifestCargo  ManifestType = "cargo"
	ManifestPython ManifestType = "python"
	ManifestNode   ManifestType = "node"
	ManifestDeno   ManifestType = "deno"
)

var manifestFilenames = map[string]ManifestType{
	"cargo.toml":     ManifestCargo,
	"pyproject.toml": ManifestPython,
	"package.json":   ManifestNode,
	"deno.json":      ManifestDeno,
}

var toolchainAliases = map[string]ManifestType{
	"cargo":          ManifestCargo,
	"rust":           ManifestCargo,
	"cargo.toml":     ManifestCargo,
	"pyproject":      ManifestPython,
	"pyproject.toml": ManifestPython,
	"python":         ManifestPython,
	"package":        ManifestNode,
	"package.json":   ManifestNode,
	"npm":            ManifestNode,
	"node":           ManifestNode,
	"deno":           ManifestDeno,
	"deno.json":      ManifestDeno,
}

var skipWalkDirs = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
	"dist":         {},
	"target":       {},
	"tmp":          {},
}

// Manifest describes a discovered manifest file with its current version.
type Manifest struct {
	Type    ManifestType
	Path    string
	RelPath string
	Version string
	Name    string
}

// DisplayLabel returns a concise label for TUI listings.
func (m Manifest) DisplayLabel() string {
	label := m.RelPath
	if m.Name != "" {
		label = fmt.Sprintf("%s · %s", label, m.Name)
	}
	if m.Version != "" {
		label = fmt.Sprintf("%s @ %s", label, m.Version)
	}
	return label
}

// Discover scans the repository tree for supported manifest files.
func Discover(root string) ([]Manifest, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var manifests []Manifest
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != absRoot {
				if _, skip := skipWalkDirs[strings.ToLower(d.Name())]; skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		kind, ok := manifestFilenames[strings.ToLower(d.Name())]
		if !ok {
			return nil
		}

		manifest, err := buildManifest(absRoot, path, kind)
		if err != nil {
			return err
		}
		manifests = append(manifests, manifest)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].RelPath < manifests[j].RelPath
	})
	return manifests, nil
}

// ResolveTargets resolves CLI selectors into manifest targets, optionally requesting a TUI selection.
func ResolveTargets(root string, selectors []string) ([]Manifest, bool, []Manifest, error) {
	if len(selectors) == 0 {
		return nil, false, nil, nil
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, false, nil, err
	}

	available, err := Discover(absRoot)
	if err != nil {
		return nil, false, nil, err
	}

	manifestByPath := make(map[string]Manifest)
	for _, manifest := range available {
		manifestByPath[filepath.Clean(manifest.Path)] = manifest
	}

	var selected []Manifest
	seen := make(map[string]struct{})
	interactive := false

	for _, raw := range selectors {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}

		lower := strings.ToLower(value)
		switch lower {
		case "interactive", "tui", "select":
			interactive = true
			continue
		}

		if kind, ok := toolchainAliases[lower]; ok {
			matched := false
			for _, manifest := range available {
				if manifest.Type == kind {
					key := filepath.Clean(manifest.Path)
					if _, exists := seen[key]; !exists {
						selected = append(selected, manifest)
						seen[key] = struct{}{}
					}
					matched = true
				}
			}
			if !matched {
				return nil, false, nil, fmt.Errorf("no %s manifest found", value)
			}
			continue
		}

		target := value
		if !filepath.IsAbs(value) {
			target = filepath.Join(absRoot, value)
		}
		target = filepath.Clean(target)

		manifest, err := loadManifest(absRoot, target)
		if err != nil {
			return nil, false, nil, err
		}

		if _, exists := seen[target]; !exists {
			selected = append(selected, manifest)
			seen[target] = struct{}{}
		}
	}

	return selected, interactive, available, nil
}

// UpdateManifest rewrites the manifest on disk with the provided version.
func UpdateManifest(manifest Manifest, newVersion string) error {
	switch manifest.Type {
	case ManifestCargo:
		return updateTomlVersion(manifest.Path, []string{"package"}, newVersion)
	case ManifestPython:
		return updateTomlVersion(manifest.Path, []string{"project", "tool.poetry"}, newVersion)
	case ManifestNode:
		return updateJSONVersion(manifest.Path, newVersion)
	case ManifestDeno:
		return updateJSONVersion(manifest.Path, newVersion)
	default:
		return fmt.Errorf("unsupported manifest type: %s", manifest.Type)
	}
}

func buildManifest(root, path string, kind ManifestType) (Manifest, error) {
	version, name, err := extractMetadata(path, kind)
	if err != nil {
		return Manifest{}, err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}

	return Manifest{
		Type:    kind,
		Path:    filepath.Clean(path),
		RelPath: filepath.Clean(rel),
		Version: version,
		Name:    name,
	}, nil
}

func loadManifest(root, path string) (Manifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("unable to read %s: %w", path, err)
	}
	if info.IsDir() {
		return Manifest{}, fmt.Errorf("%s is a directory", path)
	}

	kind, ok := manifestFilenames[strings.ToLower(filepath.Base(path))]
	if !ok {
		return Manifest{}, fmt.Errorf("unsupported toolchain file: %s", filepath.Base(path))
	}

	return buildManifest(root, path, kind)
}

func extractMetadata(path string, kind ManifestType) (string, string, error) {
	switch kind {
	case ManifestCargo:
		return parseTomlManifest(path, []string{"package"})
	case ManifestPython:
		return parseTomlManifest(path, []string{"project", "tool.poetry"})
	case ManifestNode:
		return parseJSONManifest(path)
	case ManifestDeno:
		return parseJSONManifest(path)
	default:
		return "", "", fmt.Errorf("unsupported manifest type: %s", kind)
	}
}

func parseJSONManifest(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", fmt.Errorf("failed to parse %s: %w", filepath.Base(path), err)
	}

	version, _ := payload["version"].(string)
	if version == "" {
		return "", "", fmt.Errorf("version not found in %s", filepath.Base(path))
	}

	name, _ := payload["name"].(string)
	return version, name, nil
}

func parseTomlManifest(path string, sections []string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	sectionSet := make(map[string]struct{})
	for _, section := range sections {
		sectionSet[section] = struct{}{}
	}

	var current string
	var version string
	var name string
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			current = strings.TrimSpace(strings.Trim(trimmed, "[]"))
			continue
		}
		if _, ok := sectionSet[current]; !ok {
			continue
		}

		if value, ok := parseTomlAssignment(line, "version"); ok && version == "" {
			version = value
			continue
		}
		if value, ok := parseTomlAssignment(line, "name"); ok && name == "" {
			name = value
		}
	}

	if version == "" {
		return "", "", fmt.Errorf("version not found in %s", filepath.Base(path))
	}

	return version, name, nil
}

func parseTomlAssignment(line, key string) (string, bool) {
	withoutComment := strings.Split(line, "#")[0]
	parts := strings.SplitN(withoutComment, "=", 2)
	if len(parts) != 2 {
		return "", false
	}
	if strings.TrimSpace(parts[0]) != key {
		return "", false
	}

	value := strings.TrimSpace(parts[1])
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return value, true
}

var tomlVersionPattern = regexp.MustCompile(`^(\s*version\s*=\s*)(['"])([^'"]*)(['"])(.*)$`)

func updateTomlVersion(path string, sections []string, newVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sectionSet := make(map[string]struct{})
	for _, section := range sections {
		sectionSet[section] = struct{}{}
	}

	lines := strings.Split(string(data), "\n")
	current := ""
	replaced := false

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			current = strings.TrimSpace(strings.Trim(trimmed, "[]"))
			continue
		}
		if replaced {
			continue
		}
		if _, ok := sectionSet[current]; !ok {
			continue
		}
		if matches := tomlVersionPattern.FindStringSubmatch(line); matches != nil {
			lines[idx] = fmt.Sprintf("%s%s%s%s%s", matches[1], matches[2], newVersion, matches[4], matches[5])
			replaced = true
		}
	}

	if !replaced {
		return fmt.Errorf("version not found in %s", filepath.Base(path))
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func updateJSONVersion(path string, newVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	start, end, err := findRootJSONVersion(data)
	if err != nil {
		return fmt.Errorf("version not found in %s", filepath.Base(path))
	}

	var buf bytes.Buffer
	buf.Grow(len(data) - (end - start) + len(newVersion))
	buf.Write(data[:start])
	buf.WriteString(newVersion)
	buf.Write(data[end:])

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func findRootJSONVersion(data []byte) (int, int, error) {
	depth := 0
	inString := false
	escape := false
	keyStart := -1

	for i := 0; i < len(data); i++ {
		b := data[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if b == '\\' {
				escape = true
				continue
			}
			if b == '"' {
				inString = false
				keyEnd := i
				if keyStart >= 0 {
					j := i + 1
					for j < len(data) && (data[j] == ' ' || data[j] == '\t' || data[j] == '\n' || data[j] == '\r') {
						j++
					}
					if j < len(data) && data[j] == ':' {
						key := string(data[keyStart:keyEnd])
						if key == "version" && depth == 1 {
							valueStart, valueEnd, err := locateJSONString(data, j+1)
							if err != nil {
								return -1, -1, err
							}
							return valueStart, valueEnd, nil
						}
					}
				}
				keyStart = -1
			}
			continue
		}

		switch b {
		case '"':
			inString = true
			keyStart = i + 1
		case '{', '[':
			depth++
		case '}', ']':
			if depth > 0 {
				depth--
			}
		}
	}

	return -1, -1, fmt.Errorf("version key not found")
}

func locateJSONString(data []byte, start int) (int, int, error) {
	i := start
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	if i >= len(data) || data[i] != '"' {
		return -1, -1, fmt.Errorf("version value must be a string")
	}
	valueStart := i + 1
	for j := valueStart; j < len(data); j++ {
		if data[j] == '\\' {
			j++
			continue
		}
		if data[j] == '"' {
			return valueStart, j, nil
		}
	}
	return -1, -1, fmt.Errorf("unterminated version string")
}
