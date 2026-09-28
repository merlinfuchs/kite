#!/usr/bin/env bash
# Builds kite-service for the server, swaps the binary and restarts it.
# Pass --rollback to restart with the previous binary instead.
source "$(dirname "$0")/common.sh"

# Root doesn't need sudo, which may not even be installed.
restart="if [ \"\$(id -u)\" -eq 0 ]; then systemctl restart '$SERVICE_NAME'; else sudo systemctl restart '$SERVICE_NAME'; fi
  systemctl is-active '$SERVICE_NAME'"

if [ "${1:-}" = "--rollback" ]; then
  ssh -t "$DEPLOY_HOST" "set -e
  cd '$SERVICE_DIR'
  mv kite-service.prev kite-service
  $restart"
  exit
fi

build="$(mktemp -d)"
trap 'rm -rf "$build"' EXIT

cd "$ROOT/kite-service"
CGO_ENABLED=0 GOOS=linux GOARCH="$SERVICE_ARCH" go build -o "$build/kite-service" .

scp "$build/kite-service" "$DEPLOY_HOST:$SERVICE_DIR/kite-service.new"
ssh -t "$DEPLOY_HOST" "set -e
  cd '$SERVICE_DIR'
  chmod +x kite-service.new
  if [ -f kite-service ]; then cp -p kite-service kite-service.prev; fi
  mv kite-service.new kite-service
  $restart"
