#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're on main branch
current_branch=$(git branch --show-current)
if [ "$current_branch" != "main" ]; then
    print_error "You must be on the main branch to create a release"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    print_error "Working directory is not clean. Please commit your changes first."
    exit 1
fi

# Get current version from git tags
current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
print_info "Current version: $current_version"

# Parse version number
if [[ $current_version =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    major=${BASH_REMATCH[1]}
    minor=${BASH_REMATCH[2]}
    patch=${BASH_REMATCH[3]}
else
    major=0
    minor=0
    patch=0
fi

# Calculate next versions
next_patch="v$major.$minor.$((patch + 1))"
next_minor="v$major.$((minor + 1)).0"
next_major="v$((major + 1)).0.0"

echo ""
print_info "Select release type:"
echo "1) Patch release (bug fixes): $next_patch"
echo "2) Minor release (new features): $next_minor"
echo "3) Major release (breaking changes): $next_major"
echo "4) Custom version"
echo "5) Cancel"

read -p "Enter your choice (1-5): " choice

case $choice in
    1)
        new_version=$next_patch
        ;;
    2)
        new_version=$next_minor
        ;;
    3)
        new_version=$next_major
        ;;
    4)
        read -p "Enter custom version (e.g., v1.2.3): " new_version
        if [[ ! $new_version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            print_error "Invalid version format. Please use vX.Y.Z format."
            exit 1
        fi
        ;;
    5)
        print_info "Release cancelled"
        exit 0
        ;;
    *)
        print_error "Invalid choice"
        exit 1
        ;;
esac

print_info "Creating release $new_version"

# Check if tag already exists
if git tag --list | grep -q "^$new_version$"; then
    print_error "Tag $new_version already exists"
    exit 1
fi

# Generate changelog
print_info "Generating changelog..."
if [ "$current_version" != "v0.0.0" ]; then
    changelog=$(git log --oneline --pretty=format:"- %s" "$current_version"..HEAD)
else
    changelog=$(git log --oneline --pretty=format:"- %s")
fi

echo ""
print_info "Changelog for $new_version:"
echo "$changelog"
echo ""

read -p "Do you want to proceed with this release? (y/N): " confirm
if [[ ! $confirm =~ ^[Yy]$ ]]; then
    print_info "Release cancelled"
    exit 0
fi

# Create and push tag
print_info "Creating tag $new_version..."
git tag -a "$new_version" -m "Release $new_version

$changelog"

print_info "Pushing tag to origin..."
git push origin "$new_version"

print_success "Release $new_version created successfully!"
print_info "GitHub Actions will now build and publish the binaries."
print_info "Check the progress at: https://github.com/ikasamt/web-tmux/actions"

echo ""
print_info "Once the release is complete, users can download binaries from:"
print_info "https://github.com/ikasamt/web-tmux/releases/tag/$new_version"