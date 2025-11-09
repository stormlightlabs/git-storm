---
title: Development
outline: deep
---

# Development

Storm is designed to be hackable: each package works on its own and can be
composed in tests or other Go programs. This document contains the guidance that
previously lived in the repository README.

## Guidance

1. **Composable:** packages such as `diff`, `gitlog`, and `tui` should expose
   standalone entry points that can be imported elsewhere.
2. **Frontmatter:** `.changes/*.md` entries follow this schema:

   ```yaml
   type: added
   scope: cli
   summary: Add changelog command
   ```

3. **Palette:** all TUIs must use the colors defined in `internal/style`.
4. **Command chaining:** every command should behave well in pipelines, e.g.

   ```sh
   storm unreleased list --json
   storm generate --since v1.2.0 --interactive
   storm release --bump patch --toolchain package.json
   ```

5. **Tests:**
   - Prefer teatest for Bubble Tea programs.
   - Use golden files for diff/changelog output when useful.
   - Spin up in-memory `go-git` repositories in unit tests.

## Notes

- Keep the workflow deterministic so releases can be derived from local files
  alone.
- TUIs should degrade gracefully when `stdin`/`stdout` are not TTYs.
- The binary should not depend on external services beyond git data already in
  the repo.

## Roadmap

| Phase | Deliverable                                    |
| ----- | ---------------------------------------------- |
| 1     | Core CLI (`generate`, `unreleased`, `release`) |
| 2     | Git integration and commit parsing             |
| 3     | Diff engine and styling                        |
| 4     | `.changes` storage and parsing                 |
| 5     | Interactive TUI                                |
| 6     | Keep a Changelog writer                        |
| 7     | Git tagging and CI integration                 |

## Conventional Commits

Storm follows the [Conventional Commits](https://www.conventionalcommits.org)
spec. Use the format `type(scope): summary` with optional body and footers.

### Structure

| Element | Format | Description |
| ------- | ------ | ----------- |
| Header | `<type>(<scope>): <description>` | Main commit line. |
| Scope | Optional, e.g. `api`, `cli`, `deps`. | Part of the codebase affected. |
| Breaking indicator | `!` after type/scope, e.g. `feat(api)!:` | Marks breaking change. |
| Body | Blank line then body text. | Explains what and why. |
| Footer | Blank line then metadata. | Issue references, `BREAKING CHANGE`, etc. |

### Types

| Type | Description |
| ---- | ----------- |
| `feat` | New feature. |
| `fix` | Bug fix. |
| `docs` | Documentation change. |
| `style` | Formatting-only change. |
| `refactor` | Structural change without new features or fixes. |
| `perf` | Performance improvement. |
| `test` | Adds or updates tests. |
| `build` | Build system or dependency change. |
| `ci` | CI config change. |
| `chore` | Tooling or config change outside src/test. |
| `revert` | Reverts a previous commit. |

### Examples

```text
feat(api): add pagination endpoint
fix(ui): correct button alignment issue
docs: update README installation instructions
perf(core): optimize user query performance
refactor: restructure payment module for clarity
style: apply consistent formatting
test(auth): add integration tests for OAuth flow
build(deps): bump dependencies to latest versions
ci: add GitHub Actions workflow for CI
chore: update .gitignore and clean up obsolete files
feat(api)!: remove support for legacy endpoints

BREAKING CHANGE: API no longer accepts XML-formatted requests.
```
