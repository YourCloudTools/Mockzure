#!/usr/bin/env bash
set -euo pipefail

# Usage: scripts/create-gh-pages-index.sh [target_directory]
# Default target directory is current directory.

TARGET_DIR=${1:-"."}
INDEX_PATH="${TARGET_DIR%/}/index.html"

mkdir -p "$TARGET_DIR"

if [ ! -f "$INDEX_PATH" ]; then
  echo "Creating $INDEX_PATH"
  cat > "$INDEX_PATH" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mockzure</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; margin: 40px; line-height: 1.6; }
    h1 { margin-bottom: 0.4rem; }
    .muted { color: #666; }
    ul { margin-top: 0.6rem; }
  </style>
  </head>
<body>
  <h1>Mockzure</h1>
  <div class="muted">GitHub Pages index</div>
  <ul>
    <li><a href="./">Home</a></li>
    <li><a href="./docs/">Docs</a></li>
    <li><a href="./coverage.html">Coverage</a></li>
    <li><a href="./compatibility.html">Compatibility</a></li>
  </ul>
</body>
</html>
EOF
else
  echo "Index already exists at $INDEX_PATH; skipping."
fi


