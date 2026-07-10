#!/bin/bash

set -eu -o pipefail

# Pre-flight checks - verify all required tools are available and configured
# before making any changes to the repository

check_command() {
    if ! command -v "$1" &>/dev/null; then
        echo "Error: $1 is not installed or not in PATH"
        exit 1
    fi
}

check_docker_auth() {
    local registry="$1"
    local name="$2"
    local pattern="$3"
    local config="${DOCKER_CONFIG:-$HOME/.docker}/config.json"

    if [ ! -r "$config" ] || ! grep -Eq "$pattern" "$config"; then
        echo "Error: Docker is not authenticated to $name."
        echo "Run 'docker login $registry' before releasing."
        exit 1
    fi
}

# Verify gh CLI is authenticated
if ! gh auth status &>/dev/null; then
    echo "Error: gh CLI is not authenticated. Run 'gh auth login' first."
    exit 1
fi

# Verify we can access this repository via gh
if ! gh repo view --json name &>/dev/null; then
    echo "Error: Cannot access repository via gh. Check your authentication and repository access."
    exit 1
fi

# Verify git can connect to the remote (catches SSH key issues, etc.)
if ! git ls-remote origin &>/dev/null; then
    echo "Error: Cannot connect to git remote. Check your git credentials/SSH keys."
    exit 1
fi

check_command perl
check_command go
check_command goreleaser
check_command docker

docker_version=$(docker --version 2>&1)
if [[ "$docker_version" != Docker\ version* ]]; then
    echo "Error: docker must be Docker, but found: $docker_version"
    echo "GoReleaser's Docker publishing expects Docker CLI output and fails with Podman."
    exit 1
fi

if ! docker_version_output=$(docker version 2>&1); then
    echo "Error: Docker is installed, but the Docker daemon is not reachable or the current user cannot access it."
    if [ -n "${DOCKER_HOST:-}" ]; then
        echo "DOCKER_HOST is set to: $DOCKER_HOST"
        echo "Unset DOCKER_HOST if this should use the default Docker daemon."
    fi
    echo "$docker_version_output"
    exit 1
fi

if ! docker buildx version &>/dev/null; then
    echo "Error: Docker Buildx is not installed or not available to the Docker CLI."
    echo "GoReleaser is configured to build Docker images with Buildx."
    exit 1
fi

check_docker_auth \
    docker.io \
    "Docker Hub" \
    '"(https://index\.docker\.io/v1/|docker\.io|registry-1\.docker\.io)"[[:space:]]*:'
check_docker_auth \
    ghcr.io \
    "GitHub Container Registry" \
    '"ghcr\.io"[[:space:]]*:'

# Check that we're not on the main branch
current_branch=$(git branch --show-current)
if [ "$current_branch" = "main" ]; then
    echo "Error: Releases should not be done directly on the main branch."
    echo "Please create a release branch and run this script from there."
    exit 1
fi

# Fetch latest changes and check that we're not behind origin/main
echo "Fetching from origin..."
git fetch origin

if ! git merge-base --is-ancestor origin/main HEAD; then
    echo "Error: Current branch is behind origin/main."
    echo "Please merge or rebase with origin/main before releasing."
    exit 1
fi

changelog=$(cat CHANGELOG.md)

if [[ -z ${GITHUB_TOKEN:-} ]]; then
    echo 'GITHUB_TOKEN must be set for goreleaser!'
    exit 1
fi

regex='
## ([0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?) \(([0-9]{4}-[0-9]{2}-[0-9]{2})\)

((.|
)*)'

if [[ ! $changelog =~ $regex ]]; then
    echo "Could not find date line in change log!"
    exit 1
fi

version="${BASH_REMATCH[1]}"
date="${BASH_REMATCH[3]}"
notes="$(echo "${BASH_REMATCH[4]}" | sed -n -E '/^## [0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?/,$!p')"

if [[ "$date" != "$(date +"%Y-%m-%d")" ]]; then
    echo "$date is not today!"
    exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
    echo ". is not clean." >&2
    exit 1
fi

tag="v$version"

if ! grep -q "^module github.com/maxmind/geoipupdate/$(echo "$tag" | cut -d . -f 1)" go.mod; then
    echo "Tag version does not match go.mod version!"
    exit 1
fi

perl -pi -e "s/(?<=Version = \").+?(?=\")/$version/g" internal/vars/version.go

echo $'\nRelease notes:'
echo "$notes"

read -r -e -p "Continue? (y/n) " ok

if [ "$ok" != "y" ]; then
    echo "Aborting"
    exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
    git commit -m "Update for $tag" -a
fi

git push

echo "Creating tag $tag"

message="$version

$notes"

git tag -a -m "$message" "$tag"

# goreleaser's `--clean' should clear out `dist', but it didn't work for me.
rm -rf dist
if ! goreleaser release --clean -f .goreleaser.yml --release-notes <(echo "$notes"); then
    git tag -d "$tag"
    exit 1
fi

git push --tags
