#!/usr/bin/env bash
# Builds kite-service for the server, applies pending migrations, swaps the
# binary and restarts all clusters. Pass --rollback to restart with the
# previous binary instead.
source "$(dirname "$0")/common.sh"

# systemctl restart waits for the migrations, so a failed migration shows up
# here. Crashes right after starting only show up in the status check.
restart="$REMOTE_SUDO
  \$S systemctl restart $SERVICE_UNITS
  sleep 5
  for unit in $SERVICE_UNITS; do
    if ! systemctl is-active --quiet \"\$unit\"; then
      systemctl status --no-pager \"\$unit\"
      exit 1
    fi
  done
  echo 'All clusters are running'"

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
  # Migrate once before restarting. Each cluster also migrates when it starts,
  # and while one of them runs a CREATE INDEX CONCURRENTLY, the others waiting
  # for the migration lock deadlock with it. The old binary keeps running in
  # the meantime, so migrations must work with it.
  ./kite-service.new database migrate postgres up
  if [ -f kite-service ]; then cp -p kite-service kite-service.prev; fi
  mv kite-service.new kite-service
  $restart"
