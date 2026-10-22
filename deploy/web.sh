#!/usr/bin/env bash
# Builds kite-web as a static export and uploads it to the server.
source "$(dirname "$0")/common.sh"

cd "$ROOT/kite-web"
pnpm install --frozen-lockfile
rm -rf out
OUTPUT=export pnpm run build
pnpm run export-node-info

# Old hashed assets are never deleted, so tabs still open on the previous
# version can load their chunks. The pages are swapped after the assets.
rsync -az --no-owner --no-group out/_next/ "$DEPLOY_HOST:$WEB_DIR/_next/"
rsync -az --no-owner --no-group --delete --exclude /_next/ out/ "$DEPLOY_HOST:$WEB_DIR/"
