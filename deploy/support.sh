#!/usr/bin/env bash
# Builds the knowledge from the current docs and block catalog, builds
# kite-support for the server with it embedded, swaps the binary and restarts
# the bot. Pass --rollback to restart with the previous binary instead.
source "$(dirname "$0")/common.sh"

restart="$REMOTE_SUDO
  \$S systemctl restart kite-support
  sleep 5
  if ! systemctl is-active --quiet kite-support; then
    systemctl status --no-pager kite-support
    exit 1
  fi
  echo 'kite-support is running'"

if [ "${1:-}" = "--rollback" ]; then
  ssh -t "$DEPLOY_HOST" "set -e
  cd '$SUPPORT_DIR'
  mv kite-support.prev kite-support
  $restart"
  exit
fi

build="$(mktemp -d)"
trap 'rm -rf "$build"' EXIT

cd "$ROOT/kite-support"
go run . index
CGO_ENABLED=0 GOOS=linux GOARCH="$SERVICE_ARCH" go build -o "$build/kite-support" .

scp "$build/kite-support" "$DEPLOY_HOST:$SUPPORT_DIR/kite-support.new"
ssh -t "$DEPLOY_HOST" "set -e
  cd '$SUPPORT_DIR'
  chmod +x kite-support.new
  if [ -f kite-support ]; then cp -p kite-support kite-support.prev; fi
  mv kite-support.new kite-support
  $restart"
