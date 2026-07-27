#!/usr/bin/env bash
set -euo pipefail

REPORT_FILE="${1:-report.htm}"
REPORT_NAME="${2:-$(basename "$REPORT_FILE")}"
FITNESS_REPORT_URL="${FITNESS_REPORT_URL:-https://111.230.63.109:19797}"
FITNESS_CA_FILE="${FITNESS_CA_FILE:-/root/.config/fitness-tracker/tls/ca.crt}"

if [ ! -f "$REPORT_FILE" ]; then
  echo "Report file not found: $REPORT_FILE" >&2
  exit 1
fi

case "$REPORT_NAME" in
  ""|.*|*[!A-Za-z0-9._-]*)
    echo "Remote report name must use only letters, digits, dot, underscore, and hyphen" >&2
    exit 1
    ;;
esac

case "${REPORT_NAME,,}" in
  *.htm|*.html)
    ;;
  *)
    echo "Remote report name must end in .htm or .html" >&2
    exit 1
    ;;
esac

if [ -z "${REPORT_UPLOAD_TOKEN:-}" ]; then
  echo "REPORT_UPLOAD_TOKEN is required" >&2
  exit 1
fi
case "$REPORT_UPLOAD_TOKEN" in
  *[$' \t\r\n']*)
    echo "REPORT_UPLOAD_TOKEN must not contain whitespace" >&2
    exit 1
    ;;
esac

curl_args=(
  --fail
  --show-error
  --silent
  --request PUT
  --header "Content-Type: text/html; charset=utf-8"
  --data-binary "@$REPORT_FILE"
)
if [[ "$FITNESS_REPORT_URL" == https://* ]] && [ -f "$FITNESS_CA_FILE" ]; then
  curl_args+=(--cacert "$FITNESS_CA_FILE")
fi

upload_url="${FITNESS_REPORT_URL%/}/api/reports/$REPORT_NAME"
printf 'Authorization: Bearer %s\n' "$REPORT_UPLOAD_TOKEN" |
  curl "${curl_args[@]}" --header @- "$upload_url"
printf '\nUploaded report: %s\n' "$REPORT_NAME"
