# Research: GitHub CI/CD Pipeline

**Date**: 2026-01-11
**Feature**: 002-github-ci-cd

## Decisions

### D1: PR Testing Approach

**Decision**: Use `actions/setup-go@v5` with built-in caching + manual `go test`, `go vet`, `go fmt` commands

**Rationale**:
- `actions/setup-go@v5` has dependency caching enabled by default (no extra configuration)
- Direct Go commands are simpler and more transparent than wrapped actions
- Matches the constitution's test commands exactly (`go test -cover`, `go vet`, `go fmt`)

**Alternatives Considered**:
- `golangci-lint-action`: Overkill for this project - adds 40+ linters when we only need vet/fmt
- Custom caching with `actions/cache`: No longer needed, `setup-go@v5` handles this

### D2: Semantic Version Calculation

**Decision**: Use `PaulHatch/semantic-version@v5` for version calculation

**Rationale**:
- Lightweight: calculates version from commit history without external dependencies
- Supports conventional commits patterns (feat:, fix:, BREAKING CHANGE:)
- Outputs version as step output for use in subsequent steps
- No Node.js or Go runtime required (pure action)

**Alternatives Considered**:
- `go-semantic-release/action`: More features but requires Go runtime and creates artifacts we don't need
- `semantic-release/semantic-release`: Node.js-based, designed for npm packages, overkill for Go CLI
- Manual version in file: Requires developer discipline, error-prone

### D3: Cross-Compilation Strategy

**Decision**: Manual matrix build with `GOOS`/`GOARCH` environment variables

**Rationale**:
- AgentMail uses standard library only, no CGO dependencies
- Go's built-in cross-compilation is simple: just set `GOOS` and `GOARCH`
- Matrix strategy runs builds in parallel
- No external tools needed

**Configuration**:
```yaml
matrix:
  include:
    - goos: linux
      goarch: amd64
    - goos: darwin
      goarch: amd64
    - goos: darwin
      goarch: arm64
```

**Alternatives Considered**:
- GoReleaser: Full-featured but adds `.goreleaser.yml` config and external dependency
- `wangyoucao577/go-release-action`: Less control over build process

### D4: Release Creation

**Decision**: Use `softprops/action-gh-release@v2` with `generate_release_notes: true`

**Rationale**:
- Simple, focused action for GitHub Release creation
- Supports attaching multiple binary files
- Auto-generates release notes from PR titles (GitHub native feature)
- Well-maintained with 10k+ stars

**Alternatives Considered**:
- GoReleaser: More powerful but adds complexity we don't need
- GitHub CLI (`gh release create`): Requires more scripting for multi-file upload
- Manual API calls: Too low-level

### D5: Workflow Structure

**Decision**: Two separate workflow files: `test.yml` and `release.yml`

**Rationale**:
- Clear separation of concerns (testing vs releasing)
- Different triggers: PRs for testing, main branch push for release
- Different permissions: read-only for tests, write for releases
- Easier to maintain and debug independently

**Alternatives Considered**:
- Single workflow with conditional jobs: Harder to read, permission complexity
- Reusable workflows: Overkill for two simple workflows

## Action Versions (Pinned)

| Action | Version | Purpose |
|--------|---------|---------|
| `actions/checkout` | v4 | Repository checkout |
| `actions/setup-go` | v5 | Go installation with caching |
| `PaulHatch/semantic-version` | v5 | Version calculation |
| `softprops/action-gh-release` | v2 | Release creation |
| `actions/upload-artifact` | v4 | Artifact passing between jobs |
| `actions/download-artifact` | v4 | Artifact retrieval |

## Security Considerations

1. **Minimal Permissions**:
   - Test workflow: `contents: read` only
   - Release workflow: `contents: write` (required for tagging/releasing)

2. **Token Usage**:
   - Use `GITHUB_TOKEN` (auto-generated, scoped to repository)
   - No Personal Access Tokens needed

3. **Fork PRs**:
   - Tests run with read-only permissions
   - Secrets not exposed to fork PRs (GitHub default)

## Build Configuration

**Binary naming convention**: `agentmail-{os}-{arch}`
- `agentmail-linux-amd64`
- `agentmail-darwin-amd64`
- `agentmail-darwin-arm64`

**Build flags**:
- `CGO_ENABLED=0`: Static linking (no C dependencies)
- `-ldflags="-s -w"`: Strip debug info for smaller binaries
- `-ldflags="-X main.version={{version}}"`: Embed version at build time (optional)
