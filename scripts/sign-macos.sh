#!/usr/bin/env bash
# Local helper for macOS signing. Release builds use GoReleaser's notarize
# section with MACOS_SIGN_P12 / MACOS_SIGN_PASSWORD / MACOS_NOTARY_* secrets.
set -euo pipefail

binary_path="${1:-}"
if [[ -z "$binary_path" ]]; then
  echo "usage: $0 <binary-path>" >&2
  exit 1
fi

if [[ -z "${MACOS_SIGN_IDENTITY:-}" ]]; then
  echo "MACOS_SIGN_IDENTITY is not set; skipping codesign for $binary_path" >&2
  exit 0
fi

codesign --force --timestamp --options runtime \
  --sign "$MACOS_SIGN_IDENTITY" \
  "$binary_path"

echo "Signed $binary_path with $MACOS_SIGN_IDENTITY"
