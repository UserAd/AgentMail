# Quickstart: GitHub CI/CD Pipeline

## Overview

This feature adds two GitHub Actions workflows:
1. **Test Workflow** (`test.yml`): Runs on every pull request
2. **Release Workflow** (`release.yml`): Creates releases on merge to main

## How It Works

### Pull Request Flow

```
Developer opens PR → test.yml triggers → Go tests, vet, fmt run → Status reported on PR
```

The PR will show:
- ✅ Green checkmark if all checks pass
- ❌ Red X with failure details if any check fails

### Release Flow

```
PR merged to main → release.yml triggers → Version calculated → Binaries built → Release created
```

The workflow:
1. Analyzes commit messages since last release
2. Determines version bump (feat: → minor, fix: → patch, BREAKING CHANGE: → major)
3. Creates Git tag
4. Builds binaries for all platforms
5. Creates GitHub Release with binaries attached

## Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/) for automatic versioning:

| Prefix | Version Bump | Example |
|--------|--------------|---------|
| `fix:` | Patch (0.0.X) | `fix: resolve timeout in send command` |
| `feat:` | Minor (0.X.0) | `feat: add list command for mailbox` |
| `feat!:` or `BREAKING CHANGE:` | Major (X.0.0) | `feat!: change message format to binary` |

**Examples**:
```bash
git commit -m "fix: handle empty recipient list gracefully"
git commit -m "feat: add JSON output format option"
git commit -m "feat!: rename send command to post

BREAKING CHANGE: The send command is now named post for clarity"
```

## Modifying Workflows

### Add a New Test Step

Edit `.github/workflows/test.yml`, add step to the `test` job:

```yaml
- name: My new check
  run: my-command --here
```

### Change Go Version

Update both workflow files:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.22'  # Change version here
```

### Add Build Target

Edit `.github/workflows/release.yml`, add to the matrix:

```yaml
matrix:
  include:
    # ... existing targets ...
    - goos: windows
      goarch: amd64
```

### Change Binary Name

Edit the build step in release workflow:

```yaml
run: go build -o myapp-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/agentmail
```

## Triggering Releases Manually

If you need to release without a new commit:

```bash
# Create and push a tag manually
git tag v1.2.3
git push origin v1.2.3
```

The release workflow will trigger on the new tag.

## Troubleshooting

### Tests Fail on PR

1. Check the "Actions" tab on GitHub for detailed logs
2. Run tests locally: `go test -v ./...`
3. Check formatting: `gofmt -l .`
4. Run vet: `go vet ./...`

### Release Not Created

1. Verify the commit was pushed to `main` branch
2. Check if there are any new commits since last release
3. View the Actions tab for workflow run status
4. Ensure commit messages follow conventional format

### Wrong Version Bump

The version is determined by commit messages since last tag:
- If any commit has `BREAKING CHANGE:` → major bump
- Else if any commit has `feat:` → minor bump
- Otherwise → patch bump

To fix: create a new commit with the correct prefix.

## Files Created

```
.github/
└── workflows/
    ├── test.yml      # PR testing workflow
    └── release.yml   # Release workflow
```

No changes to existing Go source code.
