#!/usr/bin/env bash
# Builds kite-service for the server, swaps the binary and restarts the
# clusters one by one, so only part of the bots reconnect at a time.
# Pass --rollback to restart with the previous binary instead.
source "$(dirname "$0")/common.sh"

# Root doesn't need sudo, which may not even be installed. Stops at the first
# cluster that doesn't come back up, the others keep running the old binary.
restart="for unit in $SERVICE_UNITS; do
    echo \"Restarting \$unit\"
    if [ \"\$(id -u)\" -eq 0 ]; then systemctl restart \"\$unit\"; else sudo systemctl restart \"\$unit\"; fi
    sleep $SERVICE_RESTART_DELAY
    if ! systemctl is-active --quiet \"\$unit\"; then
      systemctl status --no-pager \"\$unit\"
      exit 1
    fi
  done"

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
