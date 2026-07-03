#!/bin/bash
# Build the Fynerisor Browser as an Android APK using `fyne package`.
#
# Requirements:
#   - the fyne command: go install fyne.io/fyne/v2/cmd/fyne@latest
#   - Android SDK + NDK, with ANDROID_HOME / ANDROID_NDK_HOME set (or the
#     defaults below adjusted to your machine)
#
# The startup (home) URL is baked in as custom app metadata because mobile apps
# receive no command-line arguments. It comes from FyneApp.toml by default;
# override it with HOME_URL below.
#
# Usage:
#   ./build-android.sh                                   # default arm64 build
#   HOME_URL=https://your.server/app ./build-android.sh  # bake in a home URL
#   TARGET=android ./build-android.sh                    # all ABIs (needs 32-bit NDK support)
#   RELEASE=1 ./build-android.sh                         # release build

set -euo pipefail

cd "$(dirname "$0")"

APP_ID="com.fynerisor.browser"
ICON="Icon.png"
TARGET="${TARGET:-android/arm64}"

export ANDROID_HOME="${ANDROID_HOME:-$HOME/android-sdk}"
export ANDROID_SDK_ROOT="${ANDROID_SDK_ROOT:-$ANDROID_HOME}"
export ANDROID_NDK_HOME="${ANDROID_NDK_HOME:-$HOME/downloads/android-ndk-r27d}"
export PATH="$ANDROID_HOME/platform-tools:$PATH"

if ! command -v fyne >/dev/null 2>&1; then
    echo "error: 'fyne' command not found. Install it with:" >&2
    echo "  go install fyne.io/fyne/v2/cmd/fyne@latest" >&2
    exit 1
fi

ARGS=(package -os "$TARGET" -app-id "$APP_ID" -icon "$ICON")

if [ -n "${HOME_URL:-}" ]; then
    ARGS+=(--metadata "HomeURL=$HOME_URL")
fi

if [ "${RELEASE:-0}" = "1" ]; then
    ARGS+=(--release)
fi

echo "Building for $TARGET (this may take a while on first compile)..."
fyne "${ARGS[@]}"

echo
echo "APK built:"
ls -1 ./*.apk
echo
echo "Install it with: ./install-android.sh"
