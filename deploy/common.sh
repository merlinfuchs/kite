#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG="$ROOT/deploy/.env"

if [ ! -f "$CONFIG" ]; then
  echo "Missing deploy/.env, copy deploy/.env.example and fill it in." >&2
  exit 1
fi

set -a
source "$CONFIG"
set +a

echo "Deploying $(git -C "$ROOT" rev-parse --short HEAD) on $(git -C "$ROOT" rev-parse --abbrev-ref HEAD)"
if [ -n "$(git -C "$ROOT" status --porcelain)" ]; then
  echo "Warning: the working tree has uncommitted changes, they will be deployed too."
fi
