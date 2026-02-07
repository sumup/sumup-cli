#!/usr/bin/env bash
set -euo pipefail

binary_path="${1:-}"
if [[ -z "$binary_path" ]]; then
  echo "usage: $0 <binary-path>" >&2
  exit 1
fi

# Placeholder script for future macOS signing and notarization.
# Expected env vars once enabled:
# - APPLE_CERT_BASE64
# - APPLE_CERT_PASSWORD
# - APPLE_TEAM_ID
# - APPLE_ID
# - APPLE_APP_SPECIFIC_PASSWORD
#
# Example (to enable later):
# security create-keychain -p "$KEYCHAIN_PASSWORD" build.keychain
# echo "$APPLE_CERT_BASE64" | base64 --decode > cert.p12
# security import cert.p12 -k build.keychain -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign
# codesign --force --timestamp --options runtime --sign "Developer ID Application: ..." "$binary_path"

echo "macOS signing is currently disabled. Skipping: $binary_path"
