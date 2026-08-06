#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
LBS_ROOT="$(dirname "$(dirname "$ROOT_DIR")")"
IMAGE="${BUILD_IMAGE:-gxsdn:build}"
ACTION="${1:-build}"

build_in_current_environment() {
  if ! command -v go >/dev/null 2>&1; then
    echo "go is required in the current build container"
    exit 1
  fi
  if ! command -v gcc >/dev/null 2>&1; then
    echo "gcc is required because go-sqlite3 uses CGO"
    exit 1
  fi

  echo "Building in current container: $(go version)"
  rm -rf "$ROOT_DIR/backend/frontend"
  mkdir -p "$ROOT_DIR/backend/frontend"
  cp -R "$ROOT_DIR/frontend/." "$ROOT_DIR/backend/frontend/"
  (
    cd "$ROOT_DIR/backend"
    go test -mod=vendor ./...
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
      go build -mod=vendor -buildvcs=false -o "$ROOT_DIR/assistant" .
  )
}

DOCKER_ARGS=(
  --rm
  --network host
  -v /root/.config/go:/root/.config/go
  -v /usr/local/go:/usr/local/go
  -v "$LBS_ROOT:/app"
  -v /root/go:/root/go
  --workdir /app/private/fitness-tracker
)

case "$ACTION" in
  build)
    if [ "${IN_DOCKER_BUILD:-0}" = "1" ] || ! command -v docker >/dev/null 2>&1; then
      build_in_current_environment
    else
      docker run "${DOCKER_ARGS[@]}" \
        -e IN_DOCKER_BUILD=1 \
        "$IMAGE" ./dockerbuild.sh build
    fi
    echo "Built: $ROOT_DIR/assistant"
    ;;
  shell)
    if ! command -v docker >/dev/null 2>&1; then
      echo "docker is not available; the current shell is already the build environment"
      exit 1
    fi
    docker run -it "${DOCKER_ARGS[@]}" "$IMAGE" /bin/bash
    ;;
  *)
    echo "Usage: $0 {build|shell}"
    exit 1
    ;;
esac
