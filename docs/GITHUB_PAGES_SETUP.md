# GitHub Pages RPM Repository Setup

This guide will help you set up automatic RPM package publishing to GitHub Pages.

## Prerequisites

- Repository hosted on GitHub
- Admin access to the repository
- Git installed locally

## Step-by-Step Setup

### Step 1: Enable GitHub Pages

1. Go to your repository on GitHub: `https://github.com/yourcloudtools/Mockzure`
2. Click on **Settings** (top navigation bar)
3. Scroll down and click on **Pages** (left sidebar)
4. Under "Build and deployment":
   - **Source**: Select "Deploy from a branch"
   - **Branch**: Select `gh-pages` and `/` (root)
   - Click **Save**

> **Note**: The `gh-pages` branch will be created automatically by the workflow on the first run.

### Step 2: Configure Workflow Permissions

1. In repository **Settings**, go to **Actions** → **General** (left sidebar)
2. Scroll to "Workflow permissions" section
3. Select **"Read and write permissions"**
4. Check the box for **"Allow GitHub Actions to create and approve pull requests"**
5. Click **Save**

### Step 3: Verify Workflow Files

Ensure these files exist in your repository:

```
.github/
└── workflows/
    ├── build-and-publish-rpm.yml   # Main workflow
    └── README.md                   # Workflow documentation
```

If not present, the workflow file should have been created by following this guide.

### Step 4: Trigger Your First Build

You have three options to trigger a build:

#### Option A: Create a Git Tag (Recommended)

```bash
# Create and push a version tag (for tracking purposes)
git tag v1.0.0
git push origin v1.0.0
```

This will:
- Trigger the workflow
- Build with timestamp version (e.g., `20241011.143055`)
- Publish to GitHub Pages
- Tag is used for release tracking, not versioning

#### Option B: Create a GitHub Release

1. Go to your repository on GitHub
2. Click **Releases** → **Create a new release**
3. Click **"Choose a tag"** → Type `v1.0.0` → Click **"Create new tag"**
4. Fill in:
   - **Release title**: `Mockzure v1.0.0`
   - **Description**: Release notes (optional)
5. Click **"Publish release"**

#### Option C: Manual Workflow Dispatch

1. Go to **Actions** tab
2. Click **"Build and Publish RPM"** workflow
3. Click **"Run workflow"** button
4. Click **"Run workflow"** (uses timestamp version)

All builds use timestamp-based versioning: `YYYYMMDD.HHMMSS`

### Step 5: Monitor the Build

1. Go to the **Actions** tab
2. Click on the running workflow
3. Watch the progress of both jobs:
   - **build-rpm**: Builds the RPM package
   - **publish-to-pages**: Publishes to GitHub Pages

The workflow typically takes 2-3 minutes to complete.

### Step 6: Verify Deployment

After the workflow completes:

1. Check the workflow output for the Pages URL:
   ```
   🚀 RPM repository published!
   📦 View at: https://yourcloudtools.github.io/Mockzure/
   ```

2. Wait 1-2 minutes for GitHub Pages to deploy

3. Visit the URL in your browser

4. You should see:
   - A beautiful landing page
   - List of available RPM packages
   - Installation instructions
   - Repository configuration

### Step 7: Test Installation

On a RHEL/Fedora/Azure Linux system:

```bash
# Add the repository
sudo tee /etc/yum.repos.d/mockzure.repo << 'EOF'
[mockzure]
name=Mockzure RPM Repository
baseurl=https://yourcloudtools.github.io/Mockzure/rpms
enabled=1
gpgcheck=0
EOF

# List available packages
dnf --disablerepo="*" --enablerepo="mockzure" list available

# Install the package
sudo dnf install mockzure
```

## Troubleshooting

### Issue: Workflow fails with "Permission denied"

**Solution**: Verify workflow permissions are set to "Read and write"
1. Settings → Actions → General
2. Workflow permissions → Read and write
3. Save

### Issue: Pages not deploying

**Solution**: Check GitHub Pages settings
1. Settings → Pages
2. Verify Source is set to "Deploy from a branch"
3. Verify Branch is set to `gh-pages` and `/` (root)

### Issue: "gh-pages branch not found"

**Solution**: The branch will be created automatically on first run
- Wait for the workflow to complete
- Check the Actions tab for any errors
- If workflow succeeded, the branch should exist

### Issue: 404 on GitHub Pages URL

**Solution**: Wait a few minutes
- GitHub Pages can take 1-5 minutes to deploy
- Check Settings → Pages for deployment status
- Look for a green checkmark next to the URL

### Issue: RPM metadata not found

**Solution**: Re-run the workflow
- Go to Actions → Select the workflow
- Click "Re-run jobs"
- The workflow will regenerate all metadata

## Workflow Behavior

### Automatic Triggers

The workflow automatically runs when:
- You push a tag matching `v*` (e.g., `v1.0.0`, `v2.1.3`)
- You create a GitHub Release

### Version Naming

All builds use timestamp-based versioning:

| Trigger Type | Version Format | Example |
|-------------|---------------|---------|
| Any trigger | Timestamp `YYYYMMDD.HHMMSS` | `mockzure-20241011.143055-1.x86_64.rpm` |

This ensures:
- Each build has a unique version
- Versions are sortable chronologically
- Automatic upgrades work correctly with `dnf update`

### Build Artifacts

Each build creates:
- `mockzure-VERSION-1.x86_64.rpm` - The RPM package
- `mockzure-latest.rpm` - Symlink to the newest version
- `repodata/` - YUM/DNF repository metadata
- `index.html` - Package browser page

### Package Retention

By default, all built packages are retained. To implement retention:

1. Edit `.github/workflows/build-and-publish-rpm.yml`
2. Add a cleanup step before committing to gh-pages:
   ```yaml
   - name: Clean old packages
     run: |
       cd rpms
       # Keep only the 5 most recent packages
       ls -t mockzure-*.rpm | grep -v "latest" | tail -n +6 | xargs rm -f
   ```

## Repository Structure

After successful deployment:

```
gh-pages branch
├── index.html                              # Landing page
└── rpms/
    ├── mockzure-1.0.0-1.x86_64.rpm        # Version 1.0.0
    ├── mockzure-20241011.143055-1.x86_64.rpm  # Timestamp build
    ├── mockzure-latest.rpm                # Symlink to newest
    └── repodata/                          # YUM repository metadata
        ├── repomd.xml
        ├── primary.xml.gz
        ├── filelists.xml.gz
        └── other.xml.gz
```

## Advanced Configuration

### Custom Domain

To use a custom domain for your RPM repository:

1. In Settings → Pages, add your custom domain
2. Update the `baseurl` in installation instructions
3. Update the workflow to use your domain in the index.html

### Package Signing

To sign your RPM packages:

1. Generate a GPG key pair
2. Add the private key as a GitHub secret
3. Modify the workflow to import and use the key
4. Update repository config to enable `gpgcheck=1`

Example:
```yaml
- name: Import GPG key
  run: |
    echo "${{ secrets.RPM_GPG_KEY }}" | gpg --import

- name: Sign RPM
  run: |
    rpm --addsign $RPM_FILE
```

### Multi-Architecture Builds

To build for multiple architectures:

```yaml
strategy:
  matrix:
    arch: [x86_64, aarch64]

- name: Build RPM package
  run: |
    rpmbuild -bb --target ${{ matrix.arch }} ...
```

## Next Steps

1. ✅ Set up GitHub Pages (you've done this!)
2. ✅ Configure workflow permissions
3. ✅ Trigger your first build
4. 📝 Update repository URL in `mockzure.spec` if needed
5. 📝 Add release notes to your GitHub Releases
6. 📝 Share the repository URL with your team

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Workflow README](.github/workflows/README.md)
- [Main Project README](../README.md)

## Support

If you encounter issues:
1. Check the **Actions** tab for detailed logs
2. Review the **Troubleshooting** section above
3. Open an issue in the repository with:
   - Workflow run URL
   - Error messages
   - Steps you've tried

---

**Congratulations!** 🎉 You now have an automated RPM repository hosted on GitHub Pages!

Your packages are available at: **https://yourcloudtools.github.io/Mockzure/**

