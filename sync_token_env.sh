#!/usr/bin/env bash

# Shared token-file helpers. Callers must define APP_DIR before sourcing.
SYNC_TOKEN_FILE="${SYNC_TOKEN_FILE:-$APP_DIR/.fitness-tracker.env}"

load_sync_token() {
  if [ -n "${SYNC_TOKEN:-}" ]; then
    export SYNC_TOKEN
    return 0
  fi

  if [ ! -f "$SYNC_TOKEN_FILE" ]; then
    return 1
  fi

  local line
  line="$(sed -n 's/^SYNC_TOKEN=//p' "$SYNC_TOKEN_FILE" | head -n 1)"
  if [ -z "$line" ] || [[ "$line" =~ [[:space:]] ]]; then
    echo "Invalid SYNC_TOKEN in $SYNC_TOKEN_FILE" >&2
    return 1
  fi

  SYNC_TOKEN="$line"
  export SYNC_TOKEN
}

save_sync_token() {
  if [ -z "${SYNC_TOKEN:-}" ] || [[ "$SYNC_TOKEN" =~ [[:space:]] ]]; then
    echo "SYNC_TOKEN must be non-empty and contain no whitespace" >&2
    return 1
  fi

  umask 077
  printf 'SYNC_TOKEN=%s\n' "$SYNC_TOKEN" > "$SYNC_TOKEN_FILE"
  chmod 600 "$SYNC_TOKEN_FILE"
}

ensure_local_sync_token() {
  if load_sync_token; then
    return 0
  fi

  if command -v openssl >/dev/null 2>&1; then
    SYNC_TOKEN="$(openssl rand -hex 32)"
  else
    SYNC_TOKEN="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"
  fi
  export SYNC_TOKEN
  save_sync_token
  echo "Generated sync token file: $SYNC_TOKEN_FILE"
}
