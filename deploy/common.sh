#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

set -a
source "$ROOT/deploy/config.env"
if [ -f "$ROOT/deploy/.env" ]; then
  source "$ROOT/deploy/.env"
fi
set +a

# Shells that load nvm lazily don't pass node and pnpm on to scripts.
if ! command -v pnpm >/dev/null && [ -s "${NVM_DIR:-$HOME/.nvm}/nvm.sh" ]; then
  source "${NVM_DIR:-$HOME/.nvm}/nvm.sh" >/dev/null
fi

if [ -z "${DEPLOY_HOST:-}" ]; then
  echo "Missing DEPLOY_HOST, copy deploy/.env.example to deploy/.env and fill it in." >&2
  exit 1
fi

SERVICE_UNITS=""
for ((i = 0; i < CLUSTER_COUNT; i++)); do
  SERVICE_UNITS="$SERVICE_UNITS kite-service@$i"
done

# Prefix for remote commands that need root. Root doesn't need sudo, which may
# not even be installed.
REMOTE_SUDO='S=; [ "$(id -u)" -eq 0 ] || S=sudo'

# Remote commands that wait until the web server answers after starting it,
# since systemctl returns as soon as the process runs.
REMOTE_WAIT_FOR_WEB="for i in \$(seq 1 30); do
    if curl -fs -o /dev/null http://127.0.0.1:$WEB_PORT/; then break; fi
    if [ \$i -eq 30 ]; then systemctl status --no-pager kite-web; exit 1; fi
    sleep 1
  done
  echo 'kite-web is running'"

echo "Deploying $(git -C "$ROOT" rev-parse --short HEAD) on $(git -C "$ROOT" rev-parse --abbrev-ref HEAD)"
if [ -n "$(git -C "$ROOT" status --porcelain)" ]; then
  echo "Warning: the working tree has uncommitted changes, they will be deployed too."
fi
