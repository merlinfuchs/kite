# Deploying

Scripts to deploy kite.onl, docs.kite.onl and api.kite.onl from a local checkout to a single server. The web app runs as a Next.js server, the docs are static files, and nginx routes API requests to the service cluster that runs the app. Deploying the web app or docs never restarts the bots.

```shell
cp deploy/.env.example deploy/.env   # once, then set DEPLOY_HOST

./deploy/web.sh                # kite.onl
./deploy/web.sh --rollback     # restart with the previous build
./deploy/docs.sh               # docs.kite.onl
./deploy/service.sh            # api.kite.onl and the bots
./deploy/service.sh --rollback # restart with the previous binary
./deploy/support.sh            # support bot, also takes --rollback
./deploy/server.sh             # nginx config and systemd units
```

They need `pnpm`, `go`, `rsync` and SSH access to the server. The server needs Node.js 20.9 or newer at `NODE_BIN` for the web app.

## Configuration

`config.env` holds the production setup: number of clusters, ports, domains and paths. `nginx.conf`, `kite-service@.service` and `kite-web.service` are generated from it by `generate.sh`. Don't edit them by hand, CI checks they're up to date.

To change the number of clusters, set `CLUSTER_COUNT`, run `./deploy/generate.sh`, commit, and run `./deploy/server.sh`. It restarts all clusters, since apps move between them, and switches nginx over to the new routing.

`service.sh` runs pending migrations once with the new binary while the old one keeps running, then restarts all clusters at once. Migrations have to work with the running version. Each cluster also migrates when it starts, which finds nothing left to do.

`web.sh` builds the Next.js server on your machine, so it refuses builds with native modules, which would be for the wrong platform. It stops the server, swaps in the new build and starts it, which takes about a second in which pages fail. If the new build doesn't answer within 30 seconds, it goes back to the previous one. nginx serves `public/` from the build and `/_next/static` from `WEB_DIR`, which keeps the assets of old builds for tabs still open on them.

## Moving the web app to the Next.js server

kite.onl used to be a static export that nginx served from `WEB_DIR`. To switch:

1. Install Node.js 20.9 or newer on the server, at `NODE_BIN`.
2. Run `./deploy/web.sh`. It uploads the server to `WEB_SERVER_DIR` without starting it, nginx keeps serving the export.
3. Run `./deploy/server.sh`. It starts `kite-web` and then points nginx at it.

To go back, run `./deploy/server.sh` from a commit before the switch, and `systemctl disable --now kite-web` on the server. Once that's no longer needed, delete everything in `WEB_DIR` except `_next`.

## Moving from the old setup

The server used to run everything from `/root`, with one hand-written unit per cluster (`kite-service-0` to `kite-service-3`) and the nginx config in `sites-available/api.kite.onl`. `move-to-opt.sh` moves that over once:

1. Copies `kite-service`, `kite.toml` and the backup scripts to `SERVICE_DIR`, `kite-support` and its config to `SUPPORT_DIR`, and `nirn-proxy` to `NIRN_PROXY_DIR`.
2. Points the backup cron jobs and the `kite-support` and `nirn-proxy` units at the new paths and restarts those two.
3. Renames the api.kite.onl nginx config to `NGINX_CONF_PATH`, stops and disables the old cluster units, then runs `server.sh` to start `kite-service@0` to `kite-service@3` and install the new nginx config.

The bots are down between stopping the old units and starting the new ones, a few seconds. Afterwards, edit `kite.toml` in `SERVICE_DIR`, not in `/root`. The files in `/root` and the old units stay around for going back: disable the `kite-service@` units and `systemctl enable --now kite-service-0 kite-service-1 kite-service-2 kite-service-3`. Delete them once the new setup has run for a while.

## Moving the docs and web app off Vercel

1. Run `./deploy/docs.sh`, add docs.kite.onl as a public hostname on the Cloudflare tunnel pointing at nginx like api.kite.onl, and replace its DNS record (pointing at Vercel).
2. Once that works, do the same for kite.onl with `./deploy/web.sh`. Add www.kite.onl to the tunnel too, nginx redirects it to kite.onl.
