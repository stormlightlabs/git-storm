---
title: Integration Testing Scenarios
updated: 2025-11-08
version: 1
---

## Feature Branch

```bash
# 1. Create feature branch
git checkout -b feature/new-auth

# 2. Make commits
git commit -m "feat(auth): add OAuth support"
git commit -m "test(auth): add OAuth tests"

# 3. Generate entries interactively
storm generate main HEAD --interactive

# 4. Validate all documented
storm check main HEAD

# 5. Review entries
storm unreleased list

# Expected: 2 entries created, check passes
```

## Release Preparation

```bash
# 1. Generate from last release
storm generate --since v1.0.0

# 2. Review what was generated
storm unreleased review

# 3. Add manual entry for non-code change
storm unreleased add --type changed --summary "Updated documentation"

# 4. Dry-run release
storm release --version 1.1.0 --dry-run

# 5. Execute release with tag
storm release --version 1.1.0 --tag --clear-changes

# 6. Verify
git tag -n9 v1.1.0
cat CHANGELOG.md

# Expected: Clean CHANGELOG, annotated tag, empty .changes/
```

## CI Pipeline Validation

```bash
# 1. Simulate PR with new commits
git checkout -b pr/fix-bug
git commit -m "fix(api): resolve rate limit bug"

# 2. CI check (should fail)
storm check main HEAD
# Exit code: 1

# 3. Create entry
storm unreleased partial HEAD

# 4. CI check (should pass)
storm check main HEAD
# Exit code: 0

# Expected: PR can be merged with confidence
```

## Rebase Handling

```bash
# 1. Create entries for commits
storm generate HEAD~3 HEAD

# 2. Rebase interactively (squash/reword)
git rebase -i HEAD~3

# 3. Regenerate (should detect rebased commits)
storm generate HEAD~2 HEAD

# Expected: Metadata updated, no duplicates
```
