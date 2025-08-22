#!/bin/bash

# Auto Release Script for sfDBTools_new
# Usage: ./release.sh [version]
# Example: ./release.sh 1.0.0

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if git is clean
check_git_clean() {
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Git working directory is not clean. Please commit or stash changes."
        git status --short
        exit 1
    fi
}

# Check if on main branch
check_main_branch() {
    local current_branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        log_warn "You are not on main/master branch. Current branch: $current_branch"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format. Use semantic versioning: X.Y.Z"
        exit 1
    fi
}

# Check if tag already exists
check_tag_exists() {
    local tag=$1
    if git tag -l | grep -q "^$tag$"; then
        log_error "Tag $tag already exists"
        log_info "Existing tags:"
        git tag -l | tail -5
        exit 1
    fi
}

# Get version from user input or generate
get_version() {
    local version=$1
    
    if [[ -z "$version" ]]; then
        # Get latest tag
        local latest_tag
        latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
        latest_tag=${latest_tag#v}  # Remove 'v' prefix if present
        
        log_info "Latest tag: v$latest_tag"
        
        # Suggest next patch version
        IFS='.' read -ra ADDR <<< "$latest_tag"
        local major=${ADDR[0]:-0}
        local minor=${ADDR[1]:-0}
        local patch=${ADDR[2]:-0}
        
        local suggested_patch="$major.$minor.$((patch + 1))"
        local suggested_minor="$major.$((minor + 1)).0"
        local suggested_major="$((major + 1)).0.0"
        
        echo ""
        echo "Suggested versions:"
        echo "  1) $suggested_patch (patch - bug fixes)"
        echo "  2) $suggested_minor (minor - new features)"
        echo "  3) $suggested_major (major - breaking changes)"
        echo "  4) Custom version"
        echo ""
        
        read -p "Select option (1-4): " choice
        
        case $choice in
            1) version=$suggested_patch ;;
            2) version=$suggested_minor ;;
            3) version=$suggested_major ;;
            4) 
                read -p "Enter custom version (X.Y.Z): " version
                ;;
            *)
                log_error "Invalid choice"
                exit 1
                ;;
        esac
    fi
    
    echo "$version"
}

# Update version in files (if version file exists)
update_version_files() {
    local version=$1
    
    # Update version.go if exists
    if [[ -f "internal/version/version.go" ]]; then
        log_step "Updating version in internal/version/version.go"
        sed -i "s/Version = \".*\"/Version = \"$version\"/" internal/version/version.go
        git add internal/version/version.go
    fi
    
    # Update other version files if they exist
    if [[ -f "VERSION" ]]; then
        echo "$version" > VERSION
        git add VERSION
    fi
}

# Run tests
run_tests() {
    if [[ -f "go.mod" ]]; then
        log_step "Running tests..."
        go test ./... || {
            log_error "Tests failed. Aborting release."
            exit 1
        }
        log_info "All tests passed"
    fi
}

# Build locally to verify
test_build() {
    log_step "Testing local build..."
    go build -o sfdbtools-test main.go || {
        log_error "Build failed. Aborting release."
        exit 1
    }
    
    # Test binary
    ./sfdbtools-test --version 2>/dev/null || ./sfdbtools-test version 2>/dev/null || {
        log_warn "Version command failed, but binary was built successfully"
    }
    
    rm -f sfdbtools-test
    log_info "Local build successful"
}

# Create and push tag
create_release() {
    local version=$1
    local tag="v$version"
    
    log_step "Creating release commit..."
    
    # Create release commit if there are staged changes
    if [[ -n $(git diff --cached --name-only) ]]; then
        git commit -m "release: $tag"
    fi
    
    log_step "Pushing to remote..."
    git push origin $(git branch --show-current)
    
    log_step "Creating tag: $tag"
    git tag -a "$tag" -m "Release $tag"
    
    log_step "Pushing tag: $tag"
    git push origin "$tag"
    
    log_info "Release $tag created successfully!"
    log_info "GitHub Actions will now build and publish the release."
    log_info ""
    log_info "Monitor the build at:"
    log_info "  https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
    log_info ""
    log_info "Release will be available at:"
    log_info "  https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases/tag/$tag"
}

# Generate release notes
generate_release_notes() {
    local version=$1
    local tag="v$version"
    local previous_tag
    
    previous_tag=$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "")
    
    if [[ -n "$previous_tag" ]]; then
        log_info "Changes since $previous_tag:"
        echo ""
        git log --oneline "$previous_tag..HEAD" | head -20
        echo ""
    fi
    
    log_info "Release notes will be auto-generated by GitHub Actions"
}

# Main function
main() {
    local version=$1
    
    log_info "sfDBTools_new Release Script"
    log_info "======================="
    echo ""
    
    # Pre-flight checks
    log_step "Running pre-flight checks..."
    check_git_clean
    check_main_branch
    
    # Get and validate version
    version=$(get_version "$version")
    validate_version "$version"
    check_tag_exists "v$version"
    
    log_info "Preparing release: v$version"
    echo ""
    
    # Update version files
    update_version_files "$version"
    
    # Run tests and build
    run_tests
    test_build
    
    # Show summary
    echo ""
    log_info "Release Summary:"
    log_info "  Version: v$version"
    log_info "  Branch: $(git branch --show-current)"
    log_info "  Target: Linux (amd64, arm64)"
    echo ""
    
    # Confirm release
    read -p "Proceed with release? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 0
    fi
    
    # Create release
    create_release "$version"
    generate_release_notes "$version"
    
    echo ""
    log_info "🎉 Release process completed!"
    log_info ""
    log_info "Next steps:"
    log_info "  1. Wait for GitHub Actions to complete (~2-5 minutes)"
    log_info "  2. Check the release page for published binaries"
    log_info "  3. Test installation: curl -sSL https://raw.githubusercontent.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/main/install.sh | bash"
    log_info ""
}

# Run main function with all arguments
main "$@"
