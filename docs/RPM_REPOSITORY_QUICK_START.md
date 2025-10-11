# RPM Repository Quick Start

A one-page guide to get your RPM repository up and running in 5 minutes.

## Initial Setup (One-Time)

### 1. Enable GitHub Pages (2 minutes)

```
Repository → Settings → Pages
  Source: Deploy from a branch
  Branch: gh-pages / (root)
  [Save]
```

### 2. Enable Workflow Permissions (1 minute)

```
Repository → Settings → Actions → General
  Workflow permissions: Read and write
  ☑ Allow GitHub Actions to create and approve pull requests
  [Save]
```

### 3. Trigger First Build (1 minute)

**Option A - Via Git Tag:**
```bash
git tag v1.0.0
git push origin v1.0.0
```

**Option B - Via GitHub UI:**
```
Actions → Build and Publish RPM → Run workflow
```

### 4. Wait for Deployment (1-3 minutes)

```
Actions tab → Watch workflow progress
```

### 5. Verify (30 seconds)

Visit: `https://yourcloudtools.github.io/Mockzure/`

---

## Daily Usage

### Release New Version

```bash
# Create and push a tag
git tag v1.2.3
git push origin v1.2.3

# The workflow will automatically:
# ✅ Build the RPM
# ✅ Publish to GitHub Pages
# ✅ Create download links
```

### Install on Target System

```bash
# One-time: Add repository
sudo tee /etc/yum.repos.d/mockzure.repo << 'EOF'
[mockzure]
name=Mockzure RPM Repository
baseurl=https://yourcloudtools.github.io/Mockzure/rpms
enabled=1
gpgcheck=0
EOF

# Install or update
sudo dnf install mockzure
sudo dnf update mockzure
```

### Quick Install Without Repository

```bash
# Download and install latest version
curl -LO https://yourcloudtools.github.io/Mockzure/rpms/mockzure-latest.rpm
sudo dnf install -y mockzure-latest.rpm
```

---

## Common Tasks

### Check Available Versions

```bash
dnf --disablerepo="*" --enablerepo="mockzure" list available
```

### Install Specific Version

```bash
sudo dnf install mockzure-20241011.143055
```

### View Workflow Status

```
GitHub → Actions → Latest workflow run
```

### Browse All Packages

```
https://yourcloudtools.github.io/Mockzure/
```

---

## Troubleshooting One-Liners

| Issue | Solution |
|-------|----------|
| Workflow fails | Check Actions → General → Set "Read and write" permissions |
| 404 on Pages URL | Wait 2-5 minutes for GitHub Pages to deploy |
| Can't install package | Check if Pages is enabled: Settings → Pages |
| Old package showing | Clear dnf cache: `sudo dnf clean all` |

---

## Version Format

All builds use timestamp-based versioning:

| Method | Result RPM |
|--------|------------|
| Any trigger | `mockzure-YYYYMMDD.HHMMSS-1.x86_64.rpm` |
| Example | `mockzure-20241011.143055-1.x86_64.rpm` |

**Note**: Git tags (like `v1.0.0`) are used for release tracking and triggering builds, but the RPM version is always a timestamp.

---

## Key URLs

- **RPM Repository**: `https://yourcloudtools.github.io/Mockzure/`
- **Latest RPM**: `https://yourcloudtools.github.io/Mockzure/rpms/mockzure-latest.rpm`
- **Workflow**: `.github/workflows/build-and-publish-rpm.yml`
- **Full Setup Guide**: `docs/GITHUB_PAGES_SETUP.md`

---

## File Locations (After Installation)

| Item | Path |
|------|------|
| Binary | `/usr/bin/mockzure` |
| Config | `/etc/mockzure/config.json` |
| Service | `/etc/systemd/system/mockzure.service` |
| Data | `/var/lib/mockzure/` |

---

## Post-Install Commands

```bash
# Configure (optional)
sudo nano /etc/mockzure/config.json

# Start service
sudo systemctl enable mockzure
sudo systemctl start mockzure

# Check status
sudo systemctl status mockzure

# View logs
sudo journalctl -u mockzure -f

# Test endpoint
curl http://localhost:8090/mock/azure/stats
```

---

## That's It! 🎉

You now have:
- ✅ Automated RPM builds on every release
- ✅ A YUM/DNF repository hosted on GitHub Pages
- ✅ Beautiful package browser
- ✅ Easy installation for users

**Need more details?** See `docs/GITHUB_PAGES_SETUP.md`

