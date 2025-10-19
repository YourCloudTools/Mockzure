#!/bin/bash

# Script to download artifacts from a specific GitHub Actions run
# Usage: ./scripts/download-artifact.sh <run-id> <artifact-name> <output-path>

set -e

RUN_ID="${1:-18634603463}"
ARTIFACT_NAME="${2:-coverage-report-go-1.25}"
OUTPUT_PATH="${3:-./}"

REPO="YourCloudTools/Mockzure"
TOKEN="${GITHUB_TOKEN}"

if [ -z "$TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable is required"
    exit 1
fi

echo "Downloading artifact '$ARTIFACT_NAME' from run $RUN_ID..."

# Get artifact ID
ARTIFACT_ID=$(curl -s -H "Authorization: token $TOKEN" \
    "https://api.github.com/repos/$REPO/actions/runs/$RUN_ID/artifacts" | \
    jq -r ".artifacts[] | select(.name == \"$ARTIFACT_NAME\") | .id")

if [ "$ARTIFACT_ID" = "null" ] || [ -z "$ARTIFACT_ID" ]; then
    echo "Error: Artifact '$ARTIFACT_NAME' not found for run $RUN_ID"
    echo "Available artifacts:"
    curl -s -H "Authorization: token $TOKEN" \
        "https://api.github.com/repos/$REPO/actions/runs/$RUN_ID/artifacts" | \
        jq -r ".artifacts[] | .name"
    exit 1
fi

echo "Found artifact ID: $ARTIFACT_ID"

# Download artifact
curl -L -H "Authorization: token $TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$REPO/actions/artifacts/$ARTIFACT_ID/zip" \
    -o artifact.zip

# Extract artifact
mkdir -p "$OUTPUT_PATH"
unzip -o artifact.zip -d "$OUTPUT_PATH"
rm artifact.zip

echo "Artifact downloaded successfully to $OUTPUT_PATH"
