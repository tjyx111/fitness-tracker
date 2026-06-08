#!/bin/bash

echo "Stopping any running servers..."
pkill -9 -f "go run"
pkill -9 -f "^./fitness-tracker$"

echo "Starting server..."
cd /root/lbs/private/fitness-tracker/backend
export PATH=/usr/local/go/bin:$PATH

if [ -d ../frontend ]; then
  rm -rf frontend
  mkdir -p frontend
  cp ../frontend/* frontend/
fi

# 编译并运行
go run -buildvcs=false . &
SERVER_PID=$!
echo "Server started with PID: $SERVER_PID"

# 等待服务器启动
sleep 3

# 测试API
echo ""
echo "Testing API endpoints..."
echo "1. Basic exercises API:"
curl -s "http://127.0.0.1:${PORT:-8080}/api/exercises" | python3 -c "import sys, json; d=json.load(sys.stdin); print(f'  Found {len(d)} exercises')"

echo ""
echo "2. Stats - Personal Records API:"
curl -s "http://127.0.0.1:${PORT:-8080}/api/stats/personal-records"

echo ""
echo "3. Stats - Volume API:"
curl -s "http://127.0.0.1:${PORT:-8080}/api/stats/volume?days=30"

echo ""
echo "Server is running on ${LISTEN_ADDR:-${BIND_HOST:-${HOST:-0.0.0.0}}:${PORT:-8080}}"
echo "Press Ctrl+C to stop"

wait
