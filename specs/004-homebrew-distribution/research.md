# Research: Homebrew Distribution

**Feature**: 004-homebrew-distribution
**Date**: 2026-01-12

## Research Topics

### 1. Homebrew Formula Structure for Pre-Built Binaries

**Decision**: Use conditional URL blocks with `Hardware::CPU` detection for multi-architecture support.

**Rationale**: Homebrew's Ruby DSL provides built-in architecture detection that automatically selects the correct binary for the user's Mac (Intel vs Apple Silicon). This is cleaner and more maintainable than building from source.

**Formula Pattern**:
```ruby
class Agentmail < Formula
  desc "Inter-agent communication for tmux sessions"
  homepage "https://github.com/UserAd/AgentMail"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/UserAd/AgentMail/releases/download/v0.1.1/agentmail-darwin-arm64"
      sha256 "<ARM64_SHA256>"
    end
    on_intel do
      url "https://github.com/UserAd/AgentMail/releases/download/v0.1.1/agentmail-darwin-amd64"
      sha256 "<AMD64_SHA256>"
    end
  end

  def install
    bin.install "agentmail-darwin-arm64" => "agentmail" if Hardware::CPU.arm?
    bin.install "agentmail-darwin-amd64" => "agentmail" if Hardware::CPU.intel?
  end

  test do
    # Must run outside tmux for --help to work
    assert_match "agentmail", shell_output("#{bin}/agentmail --help", 2)
  end
end
```

**Alternatives Considered**:
- Building from source with `depends_on "go"`: Slower installation, requires Go toolchain
- Single URL with runtime detection: Not supported by Homebrew's download mechanism
- Using GoReleaser: Deprecated in v2.10+, adds unnecessary complexity

### 2. Cross-Repository Formula Updates via CI

**Decision**: Use GitHub Actions with a Fine-Grained PAT to directly update the formula file in `homebrew-agentmail` repository.

**Rationale**:
- `GITHUB_TOKEN` is scoped to the current repository only and cannot modify other repos
- Fine-grained PATs allow minimal permissions (Contents: Write only on `homebrew-agentmail`)
- Direct file update is simpler than workflow_dispatch triggering

**Implementation Pattern**:
```yaml
update-homebrew:
  name: Update Homebrew Formula
  runs-on: ubuntu-latest
  needs: [version, release]
  steps:
    - name: Checkout homebrew-agentmail
      uses: actions/checkout@v4
      with:
        repository: UserAd/homebrew-agentmail
        token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
        path: homebrew-tap

    - name: Download release binaries and calculate checksums
      run: |
        curl -sL "https://github.com/UserAd/AgentMail/releases/download/v$VERSION/agentmail-darwin-amd64" -o amd64
        curl -sL "https://github.com/UserAd/AgentMail/releases/download/v$VERSION/agentmail-darwin-arm64" -o arm64
        echo "SHA256_AMD64=$(shasum -a 256 amd64 | cut -d' ' -f1)" >> $GITHUB_ENV
        echo "SHA256_ARM64=$(shasum -a 256 arm64 | cut -d' ' -f1)" >> $GITHUB_ENV

    - name: Update formula
      run: |
        # Use sed or envsubst to update version and checksums in agentmail.rb

    - name: Commit and push
      run: |
        cd homebrew-tap
        git config user.name "github-actions[bot]"
        git config user.email "github-actions[bot]@users.noreply.github.com"
        git add Formula/agentmail.rb
        git commit -m "Update agentmail to v$VERSION"
        git push
```

**Alternatives Considered**:
- `workflow_dispatch` to homebrew-agentmail: More complex, requires two workflows
- Manual updates: Doesn't meet FR-004 (automated updates)
- GitHub App: Overkill for single-repo access

**Required Secret**: `HOMEBREW_TAP_TOKEN` - Fine-grained PAT with:
- Repository access: `UserAd/homebrew-agentmail` only
- Permissions: Contents (Read and Write)

### 3. Formula Validation

**Decision**: Use `brew audit --new --formula` for formula linting before release.

**Rationale**: Homebrew provides built-in formula validation that checks for common issues before publishing.

**Validation Steps** (for local testing):
```bash
# Install from local formula
brew install --build-from-source ./Formula/agentmail.rb

# Run formula audit
brew audit --new --formula ./Formula/agentmail.rb

# Test the formula
brew test agentmail
```

### 4. Tap Repository Structure

**Decision**: Create `UserAd/homebrew-agentmail` with minimal structure.

**Structure**:
```
homebrew-agentmail/
├── Formula/
│   └── agentmail.rb
└── README.md
```

**README.md Content**:
```markdown
# homebrew-agentmail

Homebrew tap for AgentMail - Inter-agent communication for tmux sessions.

## Installation

\`\`\`bash
brew install UserAd/agentmail/agentmail
\`\`\`

Or:

\`\`\`bash
brew tap UserAd/agentmail
brew install agentmail
\`\`\`
```

### 5. Linux Support via Linuxbrew

**Decision**: Include Linux binaries in the formula using `on_linux` blocks.

**Rationale**: The release workflow already builds `linux-amd64`. Minimal additional effort to support Linuxbrew users.

**Extended Formula Pattern**:
```ruby
on_macos do
  on_arm do
    url "https://github.com/UserAd/AgentMail/releases/download/v0.1.1/agentmail-darwin-arm64"
    sha256 "<DARWIN_ARM64_SHA256>"
  end
  on_intel do
    url "https://github.com/UserAd/AgentMail/releases/download/v0.1.1/agentmail-darwin-amd64"
    sha256 "<DARWIN_AMD64_SHA256>"
  end
end

on_linux do
  url "https://github.com/UserAd/AgentMail/releases/download/v0.1.1/agentmail-linux-amd64"
  sha256 "<LINUX_AMD64_SHA256>"
end
```

## Summary of Decisions

| Topic | Decision | Impact |
|-------|----------|--------|
| Formula type | Pre-built binaries with conditional URLs | Fast installation (~30s) |
| Multi-arch | `on_macos/on_arm/on_intel` blocks | Automatic architecture detection |
| CI updates | PAT + direct file update | Automated formula updates |
| Linux support | Include `on_linux` block | Linuxbrew compatibility |
| Validation | `brew audit` in local testing | Quality assurance |

## Sources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Bottles Documentation](https://docs.brew.sh/Bottles)
- [Push commits to another repository with GitHub Actions](https://some-natalie.dev/blog/multi-repo-actions/)
- [Triggering Workflows in Another Repository](https://medium.com/hostspaceng/triggering-workflows-in-another-repository-with-github-actions-4f581f8e0ceb)
- [GitHub Actions: Use GITHUB_TOKEN with workflow_dispatch](https://github.blog/changelog/2022-09-08-github-actions-use-github_token-with-workflow_dispatch-and-repository_dispatch/)
