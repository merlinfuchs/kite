#!/usr/bin/env bash
# Builds the Next.js server of kite-web, uploads it and restarts it. Pass
# --rollback to restart the previous build instead.
source "$(dirname "$0")/common.sh"

if [ "${1:-}" = "--rollback" ]; then
  ssh -t "$DEPLOY_HOST" "set -e
  $REMOTE_SUDO
  if [ ! -d '$WEB_SERVER_DIR.prev' ]; then
    echo 'There is no previous build'
    exit 1
  fi
  \$S systemctl stop kite-web
  rm -rf '$WEB_SERVER_DIR'
  mv '$WEB_SERVER_DIR.prev' '$WEB_SERVER_DIR'
  \$S systemctl start kite-web
  $REMOTE_WAIT_FOR_WEB"
  exit
fi

cd "$ROOT/kite-web"
pnpm install --frozen-lockfile
rm -rf .next
pnpm run build
# nginx serves public/ from the server's directory, so it changes with the build.
# The build already has public/logo.svg for the preview images, so the files
# are copied into it instead of public/ being copied into it as a folder.
cp -R public/. .next/standalone/public

# The server is built here, so a native module would be for this machine and
# not the server.
natives="$(find .next/standalone -name '*.node')"
if [ -n "$natives" ]; then
  echo "The build contains native modules, which won't run on the server:" >&2
  echo "$natives" >&2
  exit 1
fi

# Starts from hard links to the running build, rsync replaces changed files
# with new ones. Every build gives all files a new mtime, so they're compared
# by checksum.
installed="$(ssh "$DEPLOY_HOST" "set -e
  rm -rf '$WEB_SERVER_DIR.new'
  if [ -d '$WEB_SERVER_DIR' ]; then cp -al '$WEB_SERVER_DIR' '$WEB_SERVER_DIR.new'; fi
  if systemctl cat kite-web >/dev/null 2>&1; then echo yes; fi")"

rsync -az --checksum --no-owner --no-group --delete .next/standalone/ "$DEPLOY_HOST:$WEB_SERVER_DIR.new/"

# Old hashed assets are never deleted, so tabs still open on the previous
# version can load their chunks. They're uploaded before the server restarts,
# and never change, since their names contain hashes of their contents.
rsync -az --ignore-existing --no-owner --no-group .next/static/ "$DEPLOY_HOST:$WEB_DIR/_next/static/"

# The server is stopped before swapping, so it never loads files of the other
# build. The previous build stays in WEB_SERVER_DIR.prev for --rollback.
ssh -t "$DEPLOY_HOST" "set -e
  $REMOTE_SUDO
  # The server runs as a user of its own and nginx reads public/ from it.
  chmod -R go+rX '$WEB_SERVER_DIR.new'
  rm -rf '$WEB_SERVER_DIR.prev'
  if [ -n '$installed' ]; then \$S systemctl stop kite-web; fi
  if [ -d '$WEB_SERVER_DIR' ]; then mv '$WEB_SERVER_DIR' '$WEB_SERVER_DIR.prev'; fi
  mv '$WEB_SERVER_DIR.new' '$WEB_SERVER_DIR'
  if [ -z '$installed' ]; then
    echo 'kite-web is uploaded, run deploy/server.sh to start it and send pages to it'
    exit 0
  fi
  \$S systemctl start kite-web
  if ! ($REMOTE_WAIT_FOR_WEB); then
    if [ -d '$WEB_SERVER_DIR.prev' ]; then
      echo 'The new build did not start, going back to the previous one'
      \$S systemctl stop kite-web
      rm -rf '$WEB_SERVER_DIR'
      mv '$WEB_SERVER_DIR.prev' '$WEB_SERVER_DIR'
      \$S systemctl start kite-web
      $REMOTE_WAIT_FOR_WEB
    fi
    exit 1
  fi"
