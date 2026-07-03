#!/bin/bash
# Install the Fynerisor Browser APK on a connected Android device via adb.
#
# By default it installs the APK produced by `fyne package -os android`
# (Fynerisor_Browser.apk in this directory). Override with the first argument
# or the APK env var. Pass a device serial via DEVICE (see `adb devices`).
#
# Usage:
#   ./install-android.sh                       # install ./Fynerisor_Browser.apk
#   ./install-android.sh path/to/other.apk     # install a specific APK
#   DEVICE=2ab30210670b7ece ./install-android.sh
#   LAUNCH=1 ./install-android.sh              # also launch the app after install

set -euo pipefail

cd "$(dirname "$0")"

APP_ID="com.fynerisor.browser"
APK="${1:-${APK:-Fynerisor_Browser.apk}}"

# Locate adb: prefer PATH, then a standard Android SDK location.
if command -v adb >/dev/null 2>&1; then
    ADB="adb"
elif [ -x "${ANDROID_HOME:-$HOME/android-sdk}/platform-tools/adb" ]; then
    ADB="${ANDROID_HOME:-$HOME/android-sdk}/platform-tools/adb"
else
    echo "error: adb not found. Install platform-tools or set ANDROID_HOME." >&2
    exit 1
fi

if [ ! -f "$APK" ]; then
    echo "error: APK not found: $APK" >&2
    echo "Build it first with: fyne package -os android -app-id $APP_ID -icon Icon.png" >&2
    exit 1
fi

# Optional explicit device selection.
DEVICE_ARGS=()
if [ -n "${DEVICE:-}" ]; then
    DEVICE_ARGS=(-s "$DEVICE")
fi

echo "Installing $APK ..."
"$ADB" "${DEVICE_ARGS[@]}" install -r "$APK"

if [ "${LAUNCH:-0}" = "1" ]; then
    echo "Launching $APP_ID ..."
    "$ADB" "${DEVICE_ARGS[@]}" shell monkey -p "$APP_ID" -c android.intent.category.LAUNCHER 1 >/dev/null
fi

echo "Done."
