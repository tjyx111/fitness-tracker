#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ANDROID_DIR="$ROOT_DIR/android"
DIST_DIR="$ROOT_DIR/dist"
ANDROID_SDK_ROOT="${ANDROID_SDK_ROOT:-/opt/android-sdk}"
GRADLE_BIN="${GRADLE_BIN:-/opt/gradle/gradle-8.9/bin/gradle}"
SIGNING_DIR="${FITNESS_ANDROID_SIGNING_DIR:-/root/.config/fitness-tracker/android}"
KEYSTORE_FILE="${FITNESS_ANDROID_KEYSTORE:-$SIGNING_DIR/fitness-release.jks}"
PASSWORD_FILE="${FITNESS_ANDROID_PASSWORD_FILE:-$SIGNING_DIR/keystore-password.txt}"
KEY_ALIAS="${FITNESS_ANDROID_KEY_ALIAS:-fitness}"

if [ ! -x "$GRADLE_BIN" ]; then
  echo "Gradle not found: $GRADLE_BIN" >&2
  exit 1
fi
if [ ! -x "$ANDROID_SDK_ROOT/build-tools/35.0.0/apksigner" ]; then
  echo "Android build tools not found under: $ANDROID_SDK_ROOT" >&2
  exit 1
fi
if ! command -v keytool >/dev/null 2>&1; then
  echo "keytool is required" >&2
  exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required" >&2
  exit 1
fi

mkdir -p "$SIGNING_DIR" "$DIST_DIR"
chmod 700 "$SIGNING_DIR"

if [ ! -f "$PASSWORD_FILE" ]; then
  umask 077
  openssl rand -base64 -out "$PASSWORD_FILE" 24
fi
chmod 600 "$PASSWORD_FILE"
export FITNESS_ANDROID_KEYSTORE_PASSWORD="$(<"$PASSWORD_FILE")"
export FITNESS_ANDROID_KEY_PASSWORD="$FITNESS_ANDROID_KEYSTORE_PASSWORD"
export FITNESS_ANDROID_KEYSTORE="$KEYSTORE_FILE"
export FITNESS_ANDROID_KEY_ALIAS="$KEY_ALIAS"
export ANDROID_SDK_ROOT

if [ ! -f "$KEYSTORE_FILE" ]; then
  keytool -genkeypair \
    -keystore "$KEYSTORE_FILE" \
    -storepass:env FITNESS_ANDROID_KEYSTORE_PASSWORD \
    -keypass:env FITNESS_ANDROID_KEY_PASSWORD \
    -alias "$KEY_ALIAS" \
    -keyalg RSA \
    -keysize 3072 \
    -validity 36500 \
    -dname "CN=Fitness Tracker Android,O=Fitness Tracker"
fi
chmod 600 "$KEYSTORE_FILE"

"$GRADLE_BIN" --no-daemon -p "$ANDROID_DIR" clean lintRelease assembleRelease

SOURCE_APK="$ANDROID_DIR/app/build/outputs/apk/release/app-release.apk"
OUTPUT_APK="$DIST_DIR/fitness-tracker.apk"
if [ ! -f "$SOURCE_APK" ]; then
  echo "Release APK was not created: $SOURCE_APK" >&2
  exit 1
fi

install -m 0644 "$SOURCE_APK" "$OUTPUT_APK"
"$ANDROID_SDK_ROOT/build-tools/35.0.0/apksigner" verify --verbose --print-certs "$OUTPUT_APK"
sha256sum "$OUTPUT_APK"
