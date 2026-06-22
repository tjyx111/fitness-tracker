#!/usr/bin/env bash
# 健身记录应用启动脚本
set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=sync_token_env.sh
source "$APP_DIR/sync_token_env.sh"
ensure_local_sync_token

cd "$APP_DIR/backend"

# 设置Go环境
export PATH=/usr/local/go/bin:$PATH

if [ -d ../frontend ]; then
  rm -rf frontend
  mkdir -p frontend
  cp ../frontend/* frontend/
fi

# 启动服务器
echo "正在启动健身记录服务器..."
echo "数据目录: ${DATA_DIR:-$(pwd)/data}"
echo "监听地址: ${LISTEN_ADDR:-${BIND_HOST:-${HOST:-0.0.0.0}}:${PORT:-8080}}"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

go run -buildvcs=false .
