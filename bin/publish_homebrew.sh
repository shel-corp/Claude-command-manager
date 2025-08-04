#!/usr/bin/env bash

# Homebrew Publishing Script for CCM (Claude Command Manager)
# This script automates the process of publishing a new version to Homebrew

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/shel-corp/Claude-command-manager"
FORMULA_FILE="Formula/ccm.rb"
TAP_REPO="homebrew-claude"
TAP_OWNER="shel-corp"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if ! command -v gh &> /dev/null; then
        missing_deps+=("gh (GitHub CLI)")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v shasum &> /dev/null; then
        missing_deps+=("shasum")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        echo "Please install the missing dependencies and try again."
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Validate version format (semantic versioning)
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        log_error "Invalid version format: $version"
        echo "Version must follow semantic versioning (e.g., v1.2.3, 1.2.3-beta.1)"
        exit 1
    fi
}

# Check if we're in a clean git state
check_git_state() {
    log_info "Checking git repository state..."
    
    if [ -n "$(git status --porcelain)" ]; then
        log_error "Working directory is not clean. Please commit or stash changes."
        git status --short
        exit 1
    fi
    
    # Make sure we're on main branch
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "main" ]; then
        log_warning "Not on main branch (currently on: $current_branch)"
        read -p "Continue anyway? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    log_success "Git repository state is clean"
}

# Create and push git tag
create_git_tag() {
    local version=$1
    log_info "Creating git tag: $version"
    
    # Check if tag already exists
    if git tag --list | grep -q "^${version}$"; then
        log_error "Tag $version already exists"
        exit 1
    fi
    
    # Create annotated tag
    git tag -a "$version" -m "Release $version"
    
    # Push tag to origin
    git push origin "$version"
    
    log_success "Created and pushed tag: $version"
}

# Calculate SHA256 of release tarball
calculate_sha256() {
    local version=$1
    local tarball_url="${REPO_URL}/archive/refs/tags/${version}.tar.gz"
    
    log_info "Calculating SHA256 for release tarball..."
    log_info "URL: $tarball_url"
    
    # Wait a moment for GitHub to generate the tarball
    sleep 5
    
    local sha256=$(curl -sL "$tarball_url" | shasum -a 256 | cut -d' ' -f1 | tr -d '\n')
    
    if [ -z "$sha256" ]; then
        log_error "Failed to calculate SHA256. Please check if the release exists on GitHub."
        exit 1
    fi
    
    log_success "SHA256: $sha256" >&2  # Send to stderr to avoid capture
    echo "$sha256"  # Only echo the clean SHA256 to stdout
}

# Update formula file
update_formula() {
    local version=$1
    local sha256=$2
    
    log_info "Updating formula file: $FORMULA_FILE"
    
    # Remove 'v' prefix if present for URL
    local clean_version=${version#v}
    
    # Create temporary file for cross-platform compatibility
    local temp_file=$(mktemp)
    
    # Update URL and SHA256
    sed "s|url \".*\"|url \"${REPO_URL}/archive/refs/tags/${version}.tar.gz\"|" "$FORMULA_FILE" | \
    sed "s|sha256 \".*\"|sha256 \"${sha256}\"|" > "$temp_file"
    
    # Replace original file
    mv "$temp_file" "$FORMULA_FILE"
    
    log_success "Updated formula file"
}

# Test formula locally
test_formula() {
    log_info "Testing formula locally..."
    
    if [ -f "./test_formula.sh" ]; then
        chmod +x ./test_formula.sh
        ./test_formula.sh
    else
        log_warning "test_formula.sh not found, skipping local tests"
    fi
    
    # Test with Homebrew if available
    if command -v brew &> /dev/null; then
        log_info "Testing with Homebrew..."
        brew install --build-from-source "./$FORMULA_FILE" || {
            log_error "Homebrew test failed"
            exit 1
        }
        brew uninstall ccm || true
    else
        log_warning "Homebrew not available, skipping brew test"
    fi
    
    log_success "Formula tests passed"
}

# Create or update Homebrew tap repository
update_tap_repo() {
    local version=$1
    
    log_info "Updating Homebrew tap repository..."
    
    # Check if tap repo exists
    if ! gh repo view "${TAP_OWNER}/${TAP_REPO}" &> /dev/null; then
        log_info "Creating Homebrew tap repository..."
        gh repo create "${TAP_OWNER}/${TAP_REPO}" --public --description "Homebrew tap for Claude Command Manager"
        
        # Clone and set up the tap repo
        git clone "https://github.com/${TAP_OWNER}/${TAP_REPO}.git" /tmp/${TAP_REPO}
        cd /tmp/${TAP_REPO}
        
        # Create Formula directory
        mkdir -p Formula
        
        # Copy our formula
        cp "${OLDPWD}/${FORMULA_FILE}" Formula/
        
        # Initial commit
        git add .
        git commit -m "Add ccm formula v${version#v}"
        git push origin main
        
        cd "$OLDPWD"
        rm -rf /tmp/${TAP_REPO}
    else
        log_info "Updating existing tap repository..."
        
        # Clone the tap repo
        git clone "https://github.com/${TAP_OWNER}/${TAP_REPO}.git" /tmp/${TAP_REPO}
        cd /tmp/${TAP_REPO}
        
        # Update formula
        cp "${OLDPWD}/${FORMULA_FILE}" Formula/
        
        # Commit changes
        git add Formula/ccm.rb
        git commit -m "Update ccm to ${version#v}"
        git push origin main
        
        cd "$OLDPWD"
        rm -rf /tmp/${TAP_REPO}
    fi
    
    log_success "Homebrew tap updated"
}

# Create GitHub release
create_github_release() {
    local version=$1
    
    log_info "Creating GitHub release..."
    
    # Create release with gh CLI
    gh release create "$version" \
        --title "Release $version" \
        --notes "Release $version of Claude Command Manager" \
        --latest
    
    log_success "Created GitHub release: $version"
}

# Show help
show_help() {
    cat << EOF
🍺 Homebrew Publishing Script for CCM

Usage: $0 [version|--help|-h]

Arguments:
  version     Version to publish (e.g., v1.1.0, 1.1.0)
              If not provided, will prompt interactively

Options:
  -h, --help  Show this help message

Examples:
  $0 v1.1.0          # Publish version 1.1.0
  $0 1.1.0           # Publish version 1.1.0 (v prefix added automatically)
  $0                 # Interactive mode

This script will:
1. Validate version format and git state
2. Create and push a git tag
3. Calculate SHA256 of the release tarball
4. Update the Homebrew formula
5. Test the formula locally
6. Create a GitHub release
7. Update the Homebrew tap repository

EOF
}

# Main function
main() {
    # Handle help
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        show_help
        exit 0
    fi
    
    echo "🍺 Homebrew Publishing Script for CCM"
    echo "====================================="
    echo
    
    # Get version from user
    if [ $# -eq 0 ]; then
        echo "Current tags:"
        git tag --list | tail -5
        echo
        read -p "Enter new version (e.g., v1.1.0): " version
    else
        version=$1
    fi
    
    # Validate inputs
    validate_version "$version"
    
    # Add 'v' prefix if not present
    if [[ ! $version =~ ^v ]]; then
        version="v$version"
    fi
    
    echo "Publishing version: $version"
    echo
    
    # Confirm before proceeding
    read -p "Continue with publishing? [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Publishing cancelled"
        exit 0
    fi
    
    # Run all steps
    check_dependencies
    check_git_state
    create_git_tag "$version"
    
    # Wait for GitHub to process the tag
    log_info "Waiting for GitHub to process the new tag..."
    sleep 10
    
    # Capture SHA256 calculation with explicit output handling
    local sha256
    sha256=$(calculate_sha256 "$version" 2>/dev/null)
    if [ -z "$sha256" ]; then
        log_error "Failed to get SHA256 from calculate_sha256 function"
        exit 1
    fi
    
    # Debug: ensure SHA256 is clean
    sha256=$(echo -n "$sha256" | tr -d '\n\r' | tr -d '[:cntrl:]')
    
    update_formula "$version" "$sha256"
    test_formula
    
    # Commit formula changes to main repo
    git add "$FORMULA_FILE"
    git commit -m "Update formula for release $version"
    git push origin main
    
    create_github_release "$version"
    update_tap_repo "$version"
    
    echo
    log_success "🎉 Successfully published CCM $version to Homebrew!"
    echo
    echo "Users can now install with:"
    echo "  brew tap ${TAP_OWNER}/claude"
    echo "  brew install ccm"
    echo
    echo "Or update with:"
    echo "  brew upgrade ccm"
}

# Run main function with all arguments
main "$@"