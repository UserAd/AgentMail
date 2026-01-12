# Data Model: Homebrew Distribution

**Feature**: 004-homebrew-distribution
**Date**: 2026-01-12

## Overview

This feature is infrastructure/configuration focused and does not introduce new data entities to the AgentMail application. The "data" managed by this feature consists of configuration files.

## Configuration Entities

### 1. Homebrew Formula (agentmail.rb)

**Location**: `UserAd/homebrew-agentmail` repository → `Formula/agentmail.rb`

**Structure** (Ruby DSL):
| Field | Type | Description |
|-------|------|-------------|
| `desc` | String | Short description of the tool |
| `homepage` | URL | Project homepage |
| `version` | SemVer | Current release version |
| `license` | SPDX ID | License identifier (MIT) |
| `url` | URL | Download URL for binary (per-architecture) |
| `sha256` | Hex String | SHA256 checksum (64 characters) |

**State Transitions**:
- Formula is **created** when tap repository is initialized
- Formula is **updated** on each AgentMail release (version, URLs, checksums change)
- Formula is **validated** via `brew audit` before publishing

### 2. GitHub Actions Secret

**Location**: `UserAd/AgentMail` repository → Settings → Secrets

| Secret Name | Purpose |
|-------------|---------|
| `HOMEBREW_TAP_TOKEN` | Fine-grained PAT for updating homebrew-agentmail repo |

**Token Permissions**:
- Repository access: `UserAd/homebrew-agentmail`
- Contents: Read and Write

### 3. Release Assets (Existing)

**Location**: GitHub Releases → `v{version}` → Assets

| Asset Name | Platform | Architecture |
|------------|----------|--------------|
| `agentmail-darwin-amd64` | macOS | Intel x86_64 |
| `agentmail-darwin-arm64` | macOS | Apple Silicon |
| `agentmail-linux-amd64` | Linux | x86_64 |

## Relationships

```
┌─────────────────────────┐
│   AgentMail Release     │
│   (GitHub Release)      │
└───────────┬─────────────┘
            │ triggers
            ▼
┌─────────────────────────┐
│   release.yml workflow  │
│   (GitHub Actions)      │
└───────────┬─────────────┘
            │ updates via PAT
            ▼
┌─────────────────────────┐
│   agentmail.rb          │
│   (Homebrew Formula)    │
└───────────┬─────────────┘
            │ references
            ▼
┌─────────────────────────┐
│   Release Assets        │
│   (darwin-*, linux-*)   │
└─────────────────────────┘
```

## No Application Data Changes

This feature does not modify:
- Message storage (`.git/mail/*.jsonl`)
- Message format (JSON schema)
- CLI behavior or output
- Any Go code or internal data structures
