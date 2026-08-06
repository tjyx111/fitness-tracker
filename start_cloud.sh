#!/usr/bin/env bash
set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="${BIN:-$APP_DIR/assistant}"
if [ ! -x "$BIN" ] && [ -x "$APP_DIR/backend/assistant" ]; then
  BIN="$APP_DIR/backend/assistant"
elif [ ! -x "$BIN" ] && [ -x "$APP_DIR/fitness-tracker" ]; then
  BIN="$APP_DIR/fitness-tracker"
fi

DATA_DIR="${DATA_DIR:-$APP_DIR/data}"
LISTEN_ADDR="${LISTEN_ADDR:-0.0.0.0:19797}"
LOG_FILE="${LOG_FILE:-$APP_DIR/assistant.log}"
PID_FILE="${PID_FILE:-$APP_DIR/assistant.pid}"
APK_FILE="${APK_FILE:-$APP_DIR/downloads/assistant.apk}"
TLS_CERT_FILE="${TLS_CERT_FILE:-}"
TLS_KEY_FILE="${TLS_KEY_FILE:-}"
REPORT_DIR="${REPORT_DIR:-$DATA_DIR/reports}"
REPORT_UPLOAD_TOKEN="${REPORT_UPLOAD_TOKEN:-}"

start() {
  if [ ! -x "$BIN" ]; then
    echo "Binary not found or not executable: $BIN"
    exit 1
  fi

  if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "assistant is already running, pid=$(cat "$PID_FILE")"
    exit 0
  fi

  mkdir -p "$DATA_DIR"
  nohup env \
    DATA_DIR="$DATA_DIR" \
    LISTEN_ADDR="$LISTEN_ADDR" \
    APK_FILE="$APK_FILE" \
    TLS_CERT_FILE="$TLS_CERT_FILE" \
    TLS_KEY_FILE="$TLS_KEY_FILE" \
    REPORT_DIR="$REPORT_DIR" \
    REPORT_UPLOAD_TOKEN="$REPORT_UPLOAD_TOKEN" \
    "$BIN" > "$LOG_FILE" 2>&1 &
  echo $! > "$PID_FILE"
  echo "assistant started"
  echo "pid: $(cat "$PID_FILE")"
  echo "listen: $LISTEN_ADDR"
  echo "data: $DATA_DIR"
  echo "reports: $REPORT_DIR"
  echo "apk: $APK_FILE"
  if [ -n "$TLS_CERT_FILE" ]; then
    echo "protocol: https"
  else
    echo "protocol: http"
  fi
  echo "log: $LOG_FILE"
}

stop() {
  if [ ! -f "$PID_FILE" ]; then
    echo "pid file not found: $PID_FILE"
    return 0
  fi

  pid="$(cat "$PID_FILE")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    echo "assistant stopped, pid=$pid"
  else
    echo "assistant is not running"
  fi
  rm -f "$PID_FILE"
}

status() {
  if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "assistant is running, pid=$(cat "$PID_FILE")"
  else
    echo "assistant is not running"
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
