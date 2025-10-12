# CI/CD Guide for Mockzure

This guide explains how to use the GitHub Actions CI/CD pipeline to create releases and publish Docker images.

## Overview

The CI/CD pipeline automatically:
- Builds multi-platform binaries (Linux, macOS, Windows)
- Creates RPM packages for RHEL-based systems
- Builds and pushes multi-platform Docker images to GitHub Container Registry
- Creates GitHub releases with all artifacts attached

## Triggering a Release

Releases are triggered automatically when you push a git tag starting with `v`:

```bash
# Create a new tag
git tag v1.0.0

# Push the tag to GitHub
git push origin v1.0.0
```

Or create and push in one command:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

### Version Format

The version should follow semantic versioning:
- `v1.0.0` - Major release
- `v1.2.3` - Minor/patch release
- `v2.0.0-beta.1` - Pre-release

The `v` prefix is optional but recommended. It will be stripped for the actual version number used in artifacts.

## What Gets Built

When you push a tag, the pipeline creates:

### 1. Binary Releases
- `mockzure-linux-amd64` - Linux x86_64
- `mockzure-linux-arm64` - Linux ARM64
- `mockzure-darwin-amd64` - macOS Intel
- `mockzure-darwin-arm64` - macOS Apple Silicon
- `mockzure-windows-amd64.exe` - Windows x86_64

### 2. RPM Package
- `mockzure-{version}.x86_64.rpm` - For RHEL/Fedora/Azure Linux

### 3. Docker Images
Multi-platform images pushed to GitHub Container Registry:
- `ghcr.io/yourcloudtools/mockzure:latest`
- `ghcr.io/yourcloudtools/mockzure:v1.0.0` (tag with v prefix)
- `ghcr.io/yourcloudtools/mockzure:1.0.0` (version without v prefix)

Supported platforms:
- `linux/amd64`
- `linux/arm64`

### 4. Source Archives
GitHub automatically includes:
- Source code (zip)
- Source code (tar.gz)

## Monitoring the Build

1. Go to your repository on GitHub
2. Click on the "Actions" tab
3. Find your workflow run (named after your tag)
4. Click on it to see the progress

The workflow has 4 main jobs:
- **Build Binaries** - Builds for all platforms in parallel
- **Build RPM** - Creates the RPM package
- **Build Docker** - Builds and pushes Docker images
- **Create Release** - Publishes the GitHub release with all artifacts

## Accessing Release Artifacts

Once the workflow completes:

### GitHub Release
Visit: `https://github.com/YourCloudTools/Mockzure/releases`

Download binaries directly from the release page.

### Docker Images
Pull images from GitHub Container Registry:

```bash
# Latest version
docker pull ghcr.io/yourcloudtools/mockzure:latest

# Specific version
docker pull ghcr.io/yourcloudtools/mockzure:v1.0.0
```

### RPM Package
Download from the release page or install directly:

```bash
# Download and install
wget https://github.com/YourCloudTools/Mockzure/releases/download/v1.0.0/mockzure-1.0.0.x86_64.rpm
sudo dnf install -y mockzure-1.0.0.x86_64.rpm
```

## Docker Image Visibility

By default, Docker images pushed to GitHub Container Registry are private. To make them public:

1. Go to your package: `https://github.com/users/YourCloudTools/packages/container/mockzure`
2. Click "Package settings"
3. Scroll to "Danger Zone"
4. Click "Change visibility" and select "Public"

## Testing Before Release

Before creating an official release, you can test the build locally:

### Test Binary Build
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o mockzure-linux-amd64 main.go

# macOS ARM64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o mockzure-darwin-arm64 main.go
```

### Test Docker Build
```bash
# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 -t mockzure:test --target production .
```

### Test RPM Build
```bash
cd deploy/rpm
./build-rpm.sh
```

## Workflow Permissions

The workflow uses the default `GITHUB_TOKEN` with these permissions:
- `contents: write` - Create releases and upload assets
- `packages: write` - Push Docker images to GitHub Container Registry

These are set in the workflow file and don't require additional configuration.

## Troubleshooting

### Build Fails on Go Build
- Check that `go.mod` is up to date: `go mod tidy`
- Verify the code compiles locally: `go build main.go`

### RPM Build Fails
- Ensure `deploy/rpm/mockzure.spec` exists and is valid
- Check that `deploy/systemd/mockzure.service` exists

### Docker Push Fails
- Verify GitHub Packages is enabled for your repository
- Check that the workflow has `packages: write` permission

### Release Creation Fails
- Ensure the tag follows the pattern `v*`
- Check that you don't already have a release with that tag
- Verify workflow has `contents: write` permission

## Manual Release (Alternative)

If you prefer to create releases manually without CI/CD:

```bash
# Build locally
go build -o mockzure main.go

# Create release on GitHub UI
# Upload your binary manually
```

However, using the automated workflow is recommended for consistency and multi-platform support.

## Updating the Workflow

The workflow file is located at:
```
.github/workflows/release.yml
```

To modify the build process:
1. Edit the workflow file
2. Commit and push changes
3. Test with a new tag

## Best Practices

1. **Always test locally first** - Build and test before tagging
2. **Use semantic versioning** - Follow `vMAJOR.MINOR.PATCH` format
3. **Write release notes** - The workflow generates basic notes, but you can edit them after
4. **Don't delete tags** - If you need to fix a release, create a new patch version
5. **Monitor the workflow** - Check that all jobs complete successfully

## Example Release Process

Complete workflow for creating a release:

```bash
# 1. Make sure you're on main and up to date
git checkout main
git pull origin main

# 2. Test the build locally
go build -o mockzure main.go
./mockzure

# 3. Test Docker build
docker build -t mockzure:test --target production .

# 4. Everything works? Create and push the tag
git tag v1.2.0
git push origin v1.2.0

# 5. Monitor the workflow on GitHub Actions
# https://github.com/YourCloudTools/Mockzure/actions

# 6. Once complete, verify the release
# https://github.com/YourCloudTools/Mockzure/releases

# 7. Test the published Docker image
docker pull ghcr.io/yourcloudtools/mockzure:v1.2.0
docker run -p 8090:8090 ghcr.io/yourcloudtools/mockzure:v1.2.0
```

## Support

For issues with the CI/CD pipeline:
1. Check the Actions logs on GitHub
2. Review this guide
3. Open an issue with the `ci/cd` label

