#!/bin/bash
set -e

# Mockzure RPM Build Script
# Builds an RPM package with timestamp-based versioning

# Determine project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RPM_DIR="$SCRIPT_DIR"

echo "🏗️  Building Mockzure RPM Package"
echo "📂 Project root: $PROJECT_ROOT"

# Check if rpmbuild is available
if ! command -v rpmbuild &> /dev/null; then
    echo "⚠️  Warning: rpmbuild not found"
    echo ""
    echo "To install rpmbuild:"
    echo "  macOS:   brew install rpm"
    echo "  RHEL:    sudo dnf install rpm-build"
    echo "  Ubuntu:  sudo apt-get install rpm"
    echo ""
    echo "Alternatively, run this script on the target Linux system."
    echo ""
fi

# Generate timestamp-based version
VERSION=$(date +%Y%m%d.%H%M%S)
echo "📦 Version: $VERSION"

# Set up RPM build environment
RPMBUILD_DIR="$HOME/rpmbuild"
mkdir -p "$RPMBUILD_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Create source tarball
echo "📦 Creating source tarball..."
cd "$PROJECT_ROOT"
TARBALL_NAME="mockzure-${VERSION}.tar.gz"

# Create temporary directory with version name
TEMP_DIR=$(mktemp -d)
SRC_DIR="$TEMP_DIR/mockzure-${VERSION}"
mkdir -p "$SRC_DIR"

# Copy source files to temporary directory
echo "  Copying source files..."
cp main.go go.mod config.json.example "$SRC_DIR/"
mkdir -p "$SRC_DIR/deploy/systemd"
cp deploy/systemd/mockzure.service "$SRC_DIR/deploy/systemd/"

# Create tarball from temporary directory
cd "$TEMP_DIR"
tar -czf "$RPMBUILD_DIR/SOURCES/$TARBALL_NAME" "mockzure-${VERSION}"

# Cleanup temporary directory
rm -rf "$TEMP_DIR"
cd "$PROJECT_ROOT"

echo "✅ Source tarball created: $TARBALL_NAME"

# Copy spec file to SPECS directory
echo "📝 Copying spec file..."
cp "$RPM_DIR/mockzure.spec" "$RPMBUILD_DIR/SPECS/"

# Build RPM
echo "🔨 Building RPM package (target: x86_64)..."
rpmbuild -bb \
    --target x86_64 \
    --define "_version $VERSION" \
    --define "_topdir $RPMBUILD_DIR" \
    --define "_unitdir /usr/lib/systemd/system" \
    "$RPMBUILD_DIR/SPECS/mockzure.spec"

# Copy built RPM to deploy/rpm directory
echo "📦 Copying RPM to deploy directory..."
RPM_FILE=$(find "$RPMBUILD_DIR/RPMS/x86_64" -name "mockzure-${VERSION}*.rpm" | head -n 1)
if [ -f "$RPM_FILE" ]; then
    cp "$RPM_FILE" "$RPM_DIR/"
    RPM_BASENAME=$(basename "$RPM_FILE")
    echo "✅ RPM package built successfully: $RPM_DIR/$RPM_BASENAME"
    echo "📊 Package size: $(du -h "$RPM_DIR/$RPM_BASENAME" | cut -f1)"
    
    # Create a symlink to latest
    cd "$RPM_DIR"
    ln -sf "$RPM_BASENAME" mockzure-latest.rpm
    echo "🔗 Symlink created: mockzure-latest.rpm"
else
    echo "❌ Error: RPM file not found"
    exit 1
fi

echo ""
echo "✅ Build complete!"
echo "📦 RPM location: $RPM_DIR/$RPM_BASENAME"

