#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLOUD_URL="${CLOUD_URL:-https://111.230.63.109:19797}"
LOCAL_DB="${LOCAL_DB:-$ROOT_DIR/backend/data/fitness.db}"
GIT_COMMIT="${GIT_COMMIT:-1}"
GIT_PUSH="${GIT_PUSH:-1}"
GIT_REMOTE="${GIT_REMOTE:-origin}"
GIT_REF="${GIT_REF:-HEAD}"
FITNESS_TLS_DIR="${FITNESS_TLS_DIR:-/root/.config/fitness-tracker/tls}"
CLOUD_CA_CERT_FILE="${CLOUD_CA_CERT_FILE:-$FITNESS_TLS_DIR/ca.crt}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required"
  exit 1
fi

if [ ! -r "$CLOUD_CA_CERT_FILE" ]; then
  echo "Required TLS CA certificate is not readable: $CLOUD_CA_CERT_FILE" >&2
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
  --cacert "$CLOUD_CA_CERT_FILE" \
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

if [ "$GIT_COMMIT" = "1" ]; then
  if ! command -v git >/dev/null 2>&1; then
    echo "git is required when GIT_COMMIT=1"
    exit 1
  fi

  cd "$ROOT_DIR"
  REL_DB="$(realpath --relative-to="$ROOT_DIR" "$LOCAL_DB")"
  git add -f "$REL_DB"

  if git diff --cached --quiet -- "$REL_DB"; then
    echo "No database changes to commit."
  else
    git commit -m "sync assistant database $(date +%Y-%m-%dT%H:%M:%S%z)" -- "$REL_DB"
    echo "Database committed to git: $REL_DB"

    if [ "$GIT_PUSH" = "1" ]; then
      git push "$GIT_REMOTE" "$GIT_REF"
      echo "Database pushed to git remote: $GIT_REMOTE $GIT_REF"
    fi
  fi
fi
