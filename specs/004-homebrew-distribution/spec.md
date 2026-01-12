# Feature Specification: Homebrew Distribution

**Feature Branch**: `004-homebrew-distribution`
**Created**: 2026-01-12
**Status**: Draft
**Input**: User description: "I want to have ability to distribute releases via homebrew for mac. It should have all required preparations and files and clear instruction in README.md"

## Clarifications

### Session 2026-01-12

- Q: How should the Homebrew formula be updated when a new AgentMail version is released? → A: Automated via CI - Release workflow automatically updates formula in homebrew-agentmail repo with new version and checksums

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install AgentMail via Homebrew (Priority: P1)

A macOS user discovers AgentMail and wants to install it quickly without building from source. They use their familiar Homebrew package manager to install the tool with a simple command.

**Why this priority**: This is the primary value proposition - making installation simple and accessible for macOS users who are accustomed to using Homebrew for CLI tools.

**Independent Test**: Can be fully tested by running `brew install UserAd/agentmail/agentmail` on a fresh macOS system and verifying the binary is available in PATH.

**Acceptance Scenarios**:

1. **Given** Homebrew is installed on macOS, **When** the user runs `brew tap UserAd/agentmail`, **Then** the tap is added successfully to Homebrew
2. **Given** the tap is added, **When** the user runs `brew install agentmail`, **Then** AgentMail is installed and available as `agentmail` command
3. **Given** Homebrew is installed, **When** the user runs `brew install UserAd/agentmail/agentmail` without pre-tapping, **Then** AgentMail is installed in a single command

---

### User Story 2 - Upgrade AgentMail via Homebrew (Priority: P2)

An existing AgentMail user installed via Homebrew wants to upgrade to the latest version when a new release is available.

**Why this priority**: Essential for ongoing maintenance and ensuring users can easily get bug fixes and new features.

**Independent Test**: Can be tested by installing an older version, then running `brew upgrade agentmail` and verifying the new version is installed.

**Acceptance Scenarios**:

1. **Given** an older version of AgentMail is installed via Homebrew, **When** a new version is released and the user runs `brew update && brew upgrade agentmail`, **Then** the latest version is installed
2. **Given** the user runs `brew update`, **When** a new AgentMail version is available, **Then** the tap formula is updated automatically

---

### User Story 3 - Find Installation Instructions in README (Priority: P3)

A potential user visits the GitHub repository and wants to understand how to install AgentMail on their Mac using Homebrew.

**Why this priority**: Documentation is crucial for discoverability and adoption, but the core functionality must exist first.

**Independent Test**: Can be tested by reading the README.md and following the documented steps to successfully install AgentMail.

**Acceptance Scenarios**:

1. **Given** a user views the README.md, **When** they look for installation instructions, **Then** they find a clearly labeled Homebrew section with step-by-step commands
2. **Given** the README contains Homebrew instructions, **When** the user copies and runs the commands, **Then** AgentMail is installed successfully

---

### Edge Cases

- What happens when the user tries to install on Linux via Homebrew (Linuxbrew)? → Formalized in FR-009: Linux binaries are supported.
- How does the system handle installation when a conflicting package named `agentmail` exists in Homebrew core? → Formalized in FR-010: Users must use the full tap path.
- What happens if the GitHub release assets are temporarily unavailable? Homebrew displays a standard download error.
- How does the formula handle architecture detection (Intel vs Apple Silicon)? → Covered by FR-002: Homebrew automatically selects the correct binary based on system architecture.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Homebrew tap shall provide a formula that installs the AgentMail binary to the user's PATH.
- **FR-002**: When the user runs `brew install UserAd/agentmail/agentmail`, the Homebrew formula shall download the appropriate pre-built binary for their macOS architecture (amd64 or arm64).
- **FR-003**: The Homebrew formula shall verify the downloaded binary using SHA256 checksums.
- **FR-004**: When a new AgentMail version is released, the CI workflow shall automatically update the tap formula with the new version number and SHA256 checksums.
- **FR-005**: The README.md shall contain a Homebrew installation section with commands for adding the tap and installing AgentMail.
- **FR-006**: The Homebrew tap repository shall follow the naming convention `homebrew-agentmail` to enable shorthand tap commands.
- **FR-007**: Where building from source is selected, the Homebrew formula shall declare Go 1.21 or later as a build dependency.
- **FR-008**: If the user attempts to install on an unsupported platform, then the Homebrew formula shall fail with a clear error message indicating supported platforms.
- **FR-009**: Where Linux is the target platform, the Homebrew formula shall download the appropriate pre-built binary for linux-amd64 architecture.
- **FR-010**: If a package name conflict exists in Homebrew core, then the user shall use the full tap path `UserAd/agentmail/agentmail` to install.

### Key Entities

- **Homebrew Tap**: A separate GitHub repository (`homebrew-agentmail`) containing formula files for Homebrew package management
- **Formula**: A Ruby file defining how to download, verify, and install AgentMail binaries
- **Release Assets**: Pre-built binaries for each supported architecture, already provided by the existing release workflow

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install AgentMail on macOS using a single Homebrew command in under 30 seconds
- **SC-002**: The formula successfully installs on both Intel (amd64) and Apple Silicon (arm64) Mac architectures
- **SC-003**: SHA256 checksum verification passes for all downloaded binaries
- **SC-004**: README.md installation instructions are complete and enable successful installation when followed step-by-step
- **SC-005**: Formula upgrades work correctly when new releases are tagged
- **SC-006**: The formula successfully installs on Linux (amd64) via Linuxbrew

## Assumptions

- The existing GitHub release workflow already produces pre-built binaries for macOS amd64 and arm64
- The user has Homebrew already installed on their system
- The GitHub repository `UserAd/homebrew-agentmail` can be created for hosting the tap
- Release binaries will continue to be published as GitHub release assets
- The formula will use binary distribution from GitHub releases for faster installation
