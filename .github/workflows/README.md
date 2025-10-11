# GitHub Actions - RPM Build and Publishing

This directory contains GitHub Actions workflows for automatically building and publishing Mockzure RPM packages to GitHub Pages.

## Workflows

### `build-and-publish-rpm.yml`

Automatically builds RPM packages and publishes them to GitHub Pages, creating a YUM/DNF repository.

**Triggers:**
- **Release tags**: When you push a tag like `v1.0.0` (builds with timestamp)
- **GitHub Releases**: When you create a release (builds with timestamp)
- **Manual dispatch**: Via GitHub Actions UI (builds with timestamp)

**Note**: All builds use timestamp-based versioning in the format `YYYYMMDD.HHMMSS` regardless of trigger type.

## Setup Instructions

### 1. Enable GitHub Pages

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Pages**
3. Under "Source", select:
   - **Source**: Deploy from a branch
   - **Branch**: `gh-pages` / `root`
4. Click **Save**

GitHub will automatically deploy your RPM repository to:
```
https://yourcloudtools.github.io/Mockzure/
```

### 2. Grant Workflow Permissions

1. Go to **Settings** → **Actions** → **General**
2. Scroll to "Workflow permissions"
3. Select **"Read and write permissions"**
4. Check **"Allow GitHub Actions to create and approve pull requests"**
5. Click **Save**

### 3. Trigger Your First Build

**Option A: Create a Release**
```bash
git tag v1.0.0
git push origin v1.0.0
```

**Option B: Create a GitHub Release**
1. Go to **Releases** → **Draft a new release**
2. Choose a tag (e.g., `v1.0.0`)
3. Fill in the release details
4. Click **Publish release**

**Option C: Manual Workflow Dispatch**
1. Go to **Actions** → **Build and Publish RPM**
2. Click **Run workflow**
3. Optionally specify a version (or leave empty for timestamp)
4. Click **Run workflow**

## How It Works

### Build Process

1. **Checkout**: Clones the repository
2. **Setup Go**: Installs Go 1.22
3. **Install RPM tools**: Installs `rpmbuild` and dependencies
4. **Determine version**: 
   - Uses tag version (e.g., `v1.0.0` → `1.0.0`)
   - Or timestamp format: `YYYYMMDD.HHMMSS`
5. **Create tarball**: Packages source files
6. **Build RPM**: Uses the spec file to build the package
7. **Upload artifact**: Stores the RPM for publishing

### Publishing Process

1. **Checkout gh-pages**: Switches to the GitHub Pages branch
2. **Copy RPM**: Adds the new package to the `rpms/` directory
3. **Create symlink**: Links `mockzure-latest.rpm` to the newest build
4. **Generate metadata**: Creates YUM/DNF repository metadata with `createrepo_c`
5. **Generate index**: Creates a beautiful HTML page to browse packages
6. **Commit and push**: Updates the `gh-pages` branch

## Using the Published Repository

### Add Repository Configuration

On your target system:

```bash
# Create repository configuration
sudo tee /etc/yum.repos.d/mockzure.repo << 'EOF'
[mockzure]
name=Mockzure RPM Repository
baseurl=https://yourcloudtools.github.io/Mockzure/rpms
enabled=1
gpgcheck=0
EOF

# Install the package
sudo dnf install mockzure

# Or install latest directly
curl -LO https://yourcloudtools.github.io/Mockzure/rpms/mockzure-latest.rpm
sudo dnf install -y mockzure-latest.rpm
```

### Installing from Repository

```bash
# List available versions
dnf --disablerepo="*" --enablerepo="mockzure" list available

# Install latest
sudo dnf install mockzure

# Install specific version
sudo dnf install mockzure-20241011.143055

# Update to latest
sudo dnf update mockzure
```

## Version Management

The workflow uses **timestamp-based versioning** for all builds:

### Timestamp Versioning
- **Format**: `YYYYMMDD.HHMMSS`
- **Example**: `mockzure-20241011.143055-1.x86_64.rpm`
- **Benefit**: Each build has a unique, sortable version number
- **Usage**: Automatic upgrades work seamlessly with `dnf update`

### Triggering Builds

Any of these methods will create a new timestamped build:

1. **Push a tag**: `git tag v1.0.0 && git push origin v1.0.0`
2. **Create a release**: Via GitHub UI
3. **Manual dispatch**: Actions → Run workflow

All methods produce: `mockzure-YYYYMMDD.HHMMSS-1.x86_64.rpm`

## Repository Structure

After publishing, your `gh-pages` branch will look like:

```
gh-pages/
├── index.html                          # Beautiful package browser
└── rpms/
    ├── mockzure-20241011.143055-1.x86_64.rpm
    ├── mockzure-20241012.120000-1.x86_64.rpm
    ├── mockzure-latest.rpm             # Symlink to newest
    └── repodata/                       # YUM/DNF metadata
        ├── repomd.xml
        ├── primary.xml.gz
        ├── filelists.xml.gz
        └── other.xml.gz
```

## Monitoring Builds

### View Workflow Runs
1. Go to the **Actions** tab in your repository
2. Click on **Build and Publish RPM**
3. View individual workflow runs and logs

### Check Published Packages
Visit your GitHub Pages site:
```
https://yourcloudtools.github.io/Mockzure/
```

## Troubleshooting

### Workflow Fails: Permission Denied
- Ensure workflow permissions are set to "Read and write"
- Check that GitHub Pages is enabled

### Packages Not Showing Up
- Wait a few minutes for GitHub Pages to deploy
- Check the workflow logs for errors
- Verify the `gh-pages` branch exists

### Repository Metadata Issues
- The workflow automatically regenerates metadata with `createrepo_c`
- If metadata is corrupted, re-run the workflow

### Build Failures
- Check that Go dependencies are up to date
- Verify the spec file is correct
- Review build logs in the Actions tab

## Advanced Usage

### Building Multiple Architectures

To build for multiple architectures (e.g., aarch64), modify the workflow:

```yaml
strategy:
  matrix:
    arch: [x86_64, aarch64]
    
- name: Build RPM package
  run: |
    rpmbuild -bb --target ${{ matrix.arch }} ...
```

### Signing Packages

To sign your RPM packages:

1. Generate a GPG key
2. Add it as a GitHub secret
3. Modify the workflow to sign packages before publishing
4. Update the repo config to enable `gpgcheck=1`

### Custom Retention

Keep only the last N packages:

```yaml
- name: Clean old packages
  run: |
    cd rpms
    ls -t mockzure-*.rpm | tail -n +6 | xargs rm -f
```

## Security Notes

- Packages are published without GPG signing (`gpgcheck=0`)
- For production use, consider implementing package signing
- GitHub Pages is public - don't include secrets in packages
- Review the configuration file handling in the spec file

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [RPM Packaging Guide](https://rpm-packaging-guide.github.io/)
- [createrepo_c Documentation](https://github.com/rpm-software-management/createrepo_c)

## Support

For issues or questions:
- Open an issue in the repository
- Check the workflow logs for detailed error messages
- Review the RPM build script in `deploy/rpm/`

