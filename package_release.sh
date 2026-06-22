#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
DIST_DIR="$ROOT_DIR/dist"
BUILD_SCRIPT="$ROOT_DIR/dockerbuild.sh"

APP_NAME="${APP_NAME:-fitness-tracker}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
PACKAGE_NAME="${PACKAGE_NAME:-${APP_NAME}-${GOOS_TARGET}-${GOARCH_TARGET}}"
PACKAGE_DIR="$DIST_DIR/$PACKAGE_NAME"
ARCHIVE_PATH="$DIST_DIR/$PACKAGE_NAME.tar.gz"

echo "Preparing package: $PACKAGE_NAME"

if [ ! -d "$BACKEND_DIR" ]; then
  echo "Backend directory not found: $BACKEND_DIR"
  exit 1
fi

if [ ! -d "$FRONTEND_DIR" ]; then
  echo "Frontend directory not found: $FRONTEND_DIR"
  exit 1
fi

if [ ! -x "$BUILD_SCRIPT" ]; then
  echo "Docker build script not found or not executable: $BUILD_SCRIPT"
  exit 1
fi

if [ "$GOOS_TARGET" != "linux" ] || [ "$GOARCH_TARGET" != "amd64" ]; then
  echo "Docker build currently supports only linux/amd64"
  exit 1
fi

echo "Building and testing in Docker..."
if [ "${SKIP_BUILD:-0}" = "1" ]; then
  if [ ! -x "$ROOT_DIR/fitness-tracker" ]; then
    echo "Existing binary not found: $ROOT_DIR/fitness-tracker"
    exit 1
  fi
  echo "Docker build skipped (SKIP_BUILD=1)"
else
  "$BUILD_SCRIPT" build
fi

echo "Creating package directory..."
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

cp "$ROOT_DIR/fitness-tracker" "$PACKAGE_DIR/$APP_NAME"
cp "$ROOT_DIR/start_cloud.sh" "$PACKAGE_DIR/start_cloud.sh"
cp "$ROOT_DIR/sync_from_cloud.sh" "$PACKAGE_DIR/sync_from_cloud.sh"
cp "$ROOT_DIR/sync_token_env.sh" "$PACKAGE_DIR/sync_token_env.sh"
chmod +x "$PACKAGE_DIR/$APP_NAME" "$PACKAGE_DIR/start_cloud.sh" "$PACKAGE_DIR/sync_from_cloud.sh"

echo "Copying data directory..."
if [ -d "$BACKEND_DIR/data" ]; then
  cp -R "$BACKEND_DIR/data" "$PACKAGE_DIR/data"
else
  mkdir -p "$PACKAGE_DIR/data"
fi

cat > "$PACKAGE_DIR/README_DEPLOY.txt" <<EOF
解压后准备:

  chmod +x $APP_NAME start_cloud.sh sync_from_cloud.sh

默认:
  二进制: ./$APP_NAME
  数据目录: ./data
  日志文件: ./fitness-tracker.log
  PID 文件: ./fitness-tracker.pid
  监听地址: 183.36.16.116:19797

覆盖监听地址:
  LISTEN_ADDR=0.0.0.0:8080 ./start_cloud.sh start

首次配置云端同步 token（使用本地 .fitness-tracker.env 中的值）:
  SYNC_TOKEN=your-local-token ./start_cloud.sh configure-token

启动云端服务（后续重启会自动读取保存的 token）:
  ./start_cloud.sh start

从另一台机器拉取数据库（先停止该机器上的服务）:
  # 自动读取脚本目录中的 .fitness-tracker.env
  ./sync_from_cloud.sh

停止:
  ./start_cloud.sh stop
EOF

echo "Creating archive..."
mkdir -p "$DIST_DIR"
tar -C "$DIST_DIR" -czf "$ARCHIVE_PATH" "$PACKAGE_NAME"

echo "Package created:"
echo "  $ARCHIVE_PATH"

if [ "${SKIP_UPLOAD:-0}" = "1" ]; then
  echo "Upload skipped (SKIP_UPLOAD=1)"
else
  echo "Uploading package..."
  curl --fail --show-error --silent \
    http://tw10b0135.onething.net:9999/files \
    --form upload=@"$ARCHIVE_PATH"
  echo "Upload completed"
fi
