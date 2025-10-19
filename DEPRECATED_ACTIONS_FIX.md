# Deprecated Actions Fix

## Issue
GitHub Actions workflows were using deprecated versions of artifact actions:
- `actions/upload-artifact@v3` (deprecated April 2024)
- `actions/download-artifact@v3` (deprecated April 2024)
- `actions/upload-release-asset@v1` (deprecated)

## Fixes Applied

### 1. Updated Upload Artifact Actions
**Files Updated**: 
- `.github/workflows/test.yml` (3 instances)
- `.github/workflows/build-and-publish-rpm.yml` (1 instance)

**Change**: `actions/upload-artifact@v3` → `actions/upload-artifact@v4`

### 2. Updated Download Artifact Actions
**Files Updated**:
- `.github/workflows/build-and-publish-rpm.yml` (1 instance)
- `.github/workflows/update-coverage-badge.yml` (1 instance)

**Change**: `actions/download-artifact@v3` → `actions/download-artifact@v4`

### 3. Updated Release Asset Upload
**File Updated**: `.github/workflows/build-and-publish-rpm.yml`

**Change**: `actions/upload-release-asset@v1` → `softprops/action-gh-release@v1`

**Before**:
```yaml
- name: Attach RPM to release
  uses: actions/upload-release-asset@v1
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  with:
    upload_url: ${{ github.event.release.upload_url }}
    asset_path: ${{ steps.rpm.outputs.RPM_FILE }}
    asset_name: ${{ steps.rpm.outputs.RPM_NAME }}
    asset_content_type: application/x-rpm
```

**After**:
```yaml
- name: Attach RPM to release
  uses: softprops/action-gh-release@v1
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  with:
    files: ${{ steps.rpm.outputs.RPM_FILE }}
```

## Benefits

1. **Future Compatibility**: Using current versions ensures workflows continue to work
2. **Security**: Latest versions include security improvements
3. **Performance**: v4 artifact actions are faster and more reliable
4. **Features**: Access to new features and bug fixes

## Verification

All workflows have been updated and syntax validated. The changes maintain the same functionality while using current, supported action versions.

## Status

✅ **COMPLETED** - All deprecated actions have been updated to current versions.
