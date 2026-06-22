#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_DIR="$ROOT_DIR"
# shellcheck source=sync_token_env.sh
source "$APP_DIR/sync_token_env.sh"
CLOUD_URL="${CLOUD_URL:-http://183.36.16.116:19797}"
LOCAL_DB="${LOCAL_DB:-$ROOT_DIR/backend/data/fitness.db}"

if ! load_sync_token; then
  echo "SYNC_TOKEN is required in the environment or $SYNC_TOKEN_FILE"
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required"
  exit 1
fi
if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required"
  exit 1
fi

mkdir -p "$(dirname "$LOCAL_DB")"

if command -v fuser >/dev/null 2>&1; then
  for file in "$LOCAL_DB" "${LOCAL_DB}-wal" "${LOCAL_DB}-shm"; do
    if [ -e "$file" ] && fuser -s "$file"; then
      echo "Local database is in use. Stop the local service before synchronizing."
      exit 1
    fi
  done
fi

TEMP_DB="$(mktemp "${LOCAL_DB}.download.XXXXXX")"
trap 'rm -f "$TEMP_DB"' EXIT

curl --fail --silent --show-error \
  --noproxy "${NO_PROXY_HOSTS:-*}" \
  --header "Authorization: Bearer $SYNC_TOKEN" \
  --output "$TEMP_DB" \
  "${CLOUD_URL%/}/api/sync/database"

CHECK_RESULT="$(sqlite3 "$TEMP_DB" 'PRAGMA integrity_check;')"
if [ "$CHECK_RESULT" != "ok" ]; then
  echo "Downloaded database failed integrity check: $CHECK_RESULT"
  exit 1
fi

if [ -f "$LOCAL_DB" ]; then
  BACKUP_DB="${LOCAL_DB}.backup.$(date +%Y%m%d-%H%M%S)"
  sqlite3 "$LOCAL_DB" ".backup '$BACKUP_DB'"
  echo "Existing local database backed up to: $BACKUP_DB"
fi

chmod 600 "$TEMP_DB"
rm -f "${LOCAL_DB}-wal" "${LOCAL_DB}-shm"
mv "$TEMP_DB" "$LOCAL_DB"
trap - EXIT
echo "Cloud database synchronized to: $LOCAL_DB"
