#!/usr/bin/env bash
# Builds kite-docs and uploads it to the server.
source "$(dirname "$0")/common.sh"

cd "$ROOT/kite-docs"
pnpm install --frozen-lockfile
pnpm run build

rsync -az --no-owner --no-group --delete build/ "$DEPLOY_HOST:$DOCS_DIR/"
