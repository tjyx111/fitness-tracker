#!/usr/bin/env bash
set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="${BIN:-$APP_DIR/fitness-tracker}"
if [ ! -x "$BIN" ] && [ -x "$APP_DIR/backend/fitness-tracker" ]; then
  BIN="$APP_DIR/backend/fitness-tracker"
fi

DATA_DIR="${DATA_DIR:-$APP_DIR/data}"
LISTEN_ADDR="${LISTEN_ADDR:-183.36.16.116:19797}"
LOG_FILE="${LOG_FILE:-$APP_DIR/fitness-tracker.log}"
PID_FILE="${PID_FILE:-$APP_DIR/fitness-tracker.pid}"

start() {
  if [ ! -x "$BIN" ]; then
    echo "Binary not found or not executable: $BIN"
    exit 1
  fi

  if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "fitness-tracker is already running, pid=$(cat "$PID_FILE")"
    exit 0
  fi

  mkdir -p "$DATA_DIR"
  nohup env DATA_DIR="$DATA_DIR" LISTEN_ADDR="$LISTEN_ADDR" "$BIN" > "$LOG_FILE" 2>&1 &
  echo $! > "$PID_FILE"
  echo "fitness-tracker started"
  echo "pid: $(cat "$PID_FILE")"
  echo "listen: $LISTEN_ADDR"
  echo "data: $DATA_DIR"
  echo "log: $LOG_FILE"
}

stop() {
  if [ ! -f "$PID_FILE" ]; then
    echo "pid file not found: $PID_FILE"
    exit 0
  fi

  pid="$(cat "$PID_FILE")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    echo "fitness-tracker stopped, pid=$pid"
  else
    echo "fitness-tracker is not running"
  fi
  rm -f "$PID_FILE"
}

status() {
  if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "fitness-tracker is running, pid=$(cat "$PID_FILE")"
  else
    echo "fitness-tracker is not running"
  fi
}

case "${1:-start}" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  restart)
    stop
    start
    ;;
  status)
    status
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac
