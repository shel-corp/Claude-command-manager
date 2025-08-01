# Homebrew Formula Setup Guide

This guide explains how to set up and use the Homebrew formula for Claude Command Manager.

## Setup Overview

We've created both manual and automated approaches for Homebrew distribution:

1. **Manual Formula** - Ready-to-use formula for immediate distribution
2. **GoReleaser Automation** - Automated releases with formula generation
3. **GitHub Actions** - Continuous integration and release automation

## Prerequisites

### For Users (Installation)
- macOS 10.15+ or Linux
- Homebrew installed: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

### For Maintainers (Development)
- Go 1.24+
- Git
- GitHub account with repository access
- GoReleaser (optional): `brew install goreleaser`

## Step 1: Create the Homebrew Tap Repository

1. **Create a new GitHub repository named `homebrew-claude-tools`**:
   ```bash
   # Clone the new repository
   git clone https://github.com/sheltontolbert/homebrew-claude-tools.git
   cd homebrew-claude-tools
   
   # Copy the formula
   mkdir -p Formula
   cp /path/to/claude_command_manager/Formula/ccm.rb Formula/
   
   # Copy the tap README
   cp /path/to/claude_command_manager/TAP_README.md README.md
   
   # Commit and push
   git add .
   git commit -m "Initial tap setup with ccm formula"
   git push origin main
   ```

## Step 2: Create Initial Release

1. **Tag and release the source project**:
   ```bash
   cd /path/to/claude_command_manager
   
   # Create and push tag
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Calculate SHA256 for the release**:
   ```bash
   # Download the release tarball
   wget https://github.com/sheltontolbert/claude_command_manager/archive/refs/tags/v1.0.0.tar.gz
   
   # Calculate SHA256
   shasum -a 256 v1.0.0.tar.gz
   ```

3. **Update the formula with the correct SHA256**:
   ```ruby
   # In Formula/ccm.rb
   sha256 "actual-calculated-sha256-here"
   ```

## Step 3: Test the Formula

### Local Testing

```bash
# Test installation from local formula
brew install --build-from-source ./Formula/ccm.rb

# Test the installed binary
ccm help

# Run formula tests
brew test ccm

# Audit the formula
brew audit --strict --online ccm

# Clean up
brew uninstall ccm
```

### Testing from Tap

```bash
# Add your tap
brew tap sheltontolbert/claude-tools

# Install from tap
brew install ccm

# Verify installation
ccm --help
```

## Step 4: Automated Releases (Optional)

### GitHub Secrets Setup

1. **Create a Personal Access Token**:
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Create token with `repo` and `workflow` scopes
   - Add as `GORELEASER_GITHUB_TOKEN` in repository secrets

2. **Repository Secrets** (in claude_command_manager repo):
   ```
   GITHUB_TOKEN (automatically provided)
   GORELEASER_GITHUB_TOKEN (your personal token)
   ```

### Automated Release Process

1. **Tag a new release**:
   ```bash
   git tag v1.1.0
   git push origin v1.1.0
   ```

2. **GitHub Actions will automatically**:
   - Build binaries for multiple platforms
   - Create GitHub release with assets
   - Update Homebrew formula in the tap
   - Submit PR to tap repository

## Step 5: Publishing to Homebrew Core (Optional)

For maximum visibility, you can submit to homebrew-core:

1. **Ensure formula meets requirements**:
   - Notable project (100+ stars, active community)
   - Stable release (no pre-release versions)
   - Comprehensive tests

2. **Submit PR to homebrew-core**:
   ```bash
   # Fork homebrew-core
   git clone https://github.com/Homebrew/homebrew-core.git
   cd homebrew-core
   
   # Create formula
   cp /path/to/Formula/ccm.rb Formula/c/
   
   # Test and audit
   brew install --build-from-source Formula/c/ccm.rb
   brew audit --strict --online ccm
   
   # Submit PR
   git checkout -b ccm
   git add Formula/c/ccm.rb
   git commit -m "ccm: new formula"
   git push origin claude-command-manager
   ```

## Usage Examples

### For End Users

```bash
# Install
brew tap sheltontolbert/claude-tools
brew install ccm

# Use
ccm
```

### For CI/CD

```bash
# In CI scripts
if ! command -v ccm &> /dev/null; then
    brew tap sheltontolbert/claude-tools
    brew install ccm
fi
```

## Maintenance

### Updating Formulas

1. **Manual updates**: Edit formula in tap repository
2. **Automated updates**: Tag new releases in source repository
3. **Version bumps**: Update VERSION file and create new tag

### Monitoring

- Watch GitHub Actions for release failures
- Monitor tap repository for issues
- Test formulas on different platforms periodically

## Troubleshooting

### Common Issues

1. **Build failures**: Check Go version compatibility
2. **SHA256 mismatches**: Recalculate checksums
3. **Missing dependencies**: Update formula dependencies
4. **Test failures**: Update test assertions

### Debug Commands

```bash
# Verbose installation
brew install --verbose ccm

# Debug formula
brew --debug install ccm

# Check formula syntax
brew audit ccm
```

---

This setup provides both immediate manual distribution and long-term automated maintenance for the Homebrew formula.