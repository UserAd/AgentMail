# Quickstart: Homebrew Distribution

**Feature**: 004-homebrew-distribution
**Date**: 2026-01-12

## Prerequisites

Before implementing this feature, ensure:

1. **GitHub Account Access**: You have admin access to create repositories under `UserAd`
2. **Existing Release Workflow**: The `release.yml` workflow is functioning and creating releases with binaries
3. **Latest Release Exists**: At least one release (e.g., `v0.1.1`) with darwin-amd64, darwin-arm64, and linux-amd64 binaries

## Implementation Steps

### Step 1: Create the Tap Repository

1. Create new GitHub repository: `UserAd/homebrew-agentmail`
   - Public repository (required for Homebrew taps)
   - Initialize with README

2. Create directory structure:
   ```bash
   mkdir -p Formula
   ```

3. Create `Formula/agentmail.rb` with initial formula (see research.md for template)

4. Update repository README with installation instructions

### Step 2: Create Fine-Grained PAT

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Create new token:
   - **Name**: `homebrew-tap-updater`
   - **Expiration**: 1 year (or custom)
   - **Repository access**: Only select repositories → `UserAd/homebrew-agentmail`
   - **Permissions**: Contents → Read and Write
3. Copy the token value

### Step 3: Add Repository Secret

1. Go to `UserAd/AgentMail` → Settings → Secrets and variables → Actions
2. Add new repository secret:
   - **Name**: `HOMEBREW_TAP_TOKEN`
   - **Value**: The PAT from Step 2

### Step 4: Update Release Workflow

Add new job to `.github/workflows/release.yml` after the `release` job:

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

    - name: Download and checksum binaries
      env:
        VERSION: ${{ needs.version.outputs.version }}
      run: |
        curl -sL "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-darwin-amd64" -o darwin-amd64
        curl -sL "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-darwin-arm64" -o darwin-arm64
        curl -sL "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-linux-amd64" -o linux-amd64

        echo "SHA_DARWIN_AMD64=$(shasum -a 256 darwin-amd64 | cut -d' ' -f1)" >> $GITHUB_ENV
        echo "SHA_DARWIN_ARM64=$(shasum -a 256 darwin-arm64 | cut -d' ' -f1)" >> $GITHUB_ENV
        echo "SHA_LINUX_AMD64=$(shasum -a 256 linux-amd64 | cut -d' ' -f1)" >> $GITHUB_ENV

    - name: Update formula
      env:
        VERSION: ${{ needs.version.outputs.version }}
      run: |
        cat > Formula/agentmail.rb << 'FORMULA'
        class Agentmail < Formula
          desc "Inter-agent communication for tmux sessions"
          homepage "https://github.com/UserAd/AgentMail"
          version "${VERSION}"
          license "MIT"

          on_macos do
            on_arm do
              url "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-darwin-arm64"
              sha256 "${SHA_DARWIN_ARM64}"
            end
            on_intel do
              url "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-darwin-amd64"
              sha256 "${SHA_DARWIN_AMD64}"
            end
          end

          on_linux do
            url "https://github.com/UserAd/AgentMail/releases/download/v${VERSION}/agentmail-linux-amd64"
            sha256 "${SHA_LINUX_AMD64}"
          end

          def install
            if OS.mac?
              bin.install "agentmail-darwin-arm64" => "agentmail" if Hardware::CPU.arm?
              bin.install "agentmail-darwin-amd64" => "agentmail" if Hardware::CPU.intel?
            else
              bin.install "agentmail-linux-amd64" => "agentmail"
            end
          end

          test do
            assert_match "agentmail", shell_output("#{bin}/agentmail --help", 2)
          end
        end
        FORMULA

        # Substitute environment variables
        sed -i "s/\${VERSION}/$VERSION/g" Formula/agentmail.rb
        sed -i "s/\${SHA_DARWIN_AMD64}/$SHA_DARWIN_AMD64/g" Formula/agentmail.rb
        sed -i "s/\${SHA_DARWIN_ARM64}/$SHA_DARWIN_ARM64/g" Formula/agentmail.rb
        sed -i "s/\${SHA_LINUX_AMD64}/$SHA_LINUX_AMD64/g" Formula/agentmail.rb

    - name: Commit and push
      env:
        VERSION: ${{ needs.version.outputs.version }}
      run: |
        git config user.name "github-actions[bot]"
        git config user.email "github-actions[bot]@users.noreply.github.com"
        git add Formula/agentmail.rb
        git commit -m "Update agentmail to v${VERSION}"
        git push
```

### Step 5: Update README.md

Add Homebrew installation section to the main README:

```markdown
### Homebrew (macOS/Linux)

```bash
brew install UserAd/agentmail/agentmail
```

Or add the tap first:

```bash
brew tap UserAd/agentmail
brew install agentmail
```
```

## Verification

### Local Testing (Before First Release)

```bash
# Clone the tap repo
git clone https://github.com/UserAd/homebrew-agentmail.git
cd homebrew-agentmail

# Install from local formula
brew install --formula ./Formula/agentmail.rb

# Verify installation
agentmail --help

# Run formula audit
brew audit --new --formula ./Formula/agentmail.rb

# Uninstall
brew uninstall agentmail
```

### Post-Release Verification

```bash
# Update Homebrew
brew update

# Install from tap
brew install UserAd/agentmail/agentmail

# Verify version matches release
agentmail --help

# Test upgrade (after next release)
brew upgrade agentmail
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `brew install` fails with 404 | Check release assets exist and URLs are correct |
| Checksum mismatch | Re-run workflow or manually calculate SHA256 |
| PAT permission denied | Verify token has Contents write access |
| Formula audit fails | Check formula syntax with `brew audit --formula` |

## Related Files

| File | Repository | Purpose |
|------|------------|---------|
| `Formula/agentmail.rb` | homebrew-agentmail | Homebrew formula |
| `.github/workflows/release.yml` | AgentMail | CI workflow with formula update |
| `README.md` | AgentMail | Installation instructions |
