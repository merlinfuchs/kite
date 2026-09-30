# Deploying

Scripts to deploy kite.onl, docs.kite.onl and api.kite.onl from a local checkout to a single server. nginx serves the web app and docs as static files, so deploying them never restarts the bots, and routes API requests to the service cluster that runs the app.

```shell
cp deploy/.env.example deploy/.env   # once, then set DEPLOY_HOST

./deploy/web.sh                # kite.onl
./deploy/docs.sh               # docs.kite.onl
./deploy/service.sh            # api.kite.onl and the bots
./deploy/service.sh --rollback # restart with the previous binary
./deploy/support.sh            # support bot, needs OPENAI_API_KEY, also takes --rollback
./deploy/server.sh             # nginx config and systemd unit
```

They need `pnpm`, `go`, `rsync` and SSH access to the server.

## Configuration

`config.env` holds the production setup: number of clusters, ports, domains and paths. `nginx.conf` and `kite-service@.service` are generated from it by `generate.sh`. Don't edit them by hand, CI checks they're up to date.

To change the number of clusters, set `CLUSTER_COUNT`, run `./deploy/generate.sh`, commit, and run `./deploy/server.sh`. It restarts all clusters, since apps move between them, and switches nginx over to the new routing.

`service.sh` restarts all clusters at once. Each one runs pending migrations before it starts, and the migrations take a lock, so they only run once.

## Moving from the old setup

The server used to run everything from `/root`, with one hand-written unit per cluster (`kite-service-0` to `kite-service-3`) and the nginx config in `sites-available/api.kite.onl`. `move-to-opt.sh` moves that over once:

1. Copies `kite-service`, `kite.toml` and the backup scripts to `SERVICE_DIR`, `kite-support` and its config to `SUPPORT_DIR`, and `nirn-proxy` to `NIRN_PROXY_DIR`.
2. Points the backup cron jobs and the `kite-support` and `nirn-proxy` units at the new paths and restarts those two.
3. Renames the api.kite.onl nginx config to `NGINX_CONF_PATH`, stops and disables the old cluster units, then runs `server.sh` to start `kite-service@0` to `kite-service@3` and install the new nginx config.

The bots are down between stopping the old units and starting the new ones, a few seconds. Afterwards, edit `kite.toml` in `SERVICE_DIR`, not in `/root`. The files in `/root` and the old units stay around for going back: disable the `kite-service@` units and `systemctl enable --now kite-service-0 kite-service-1 kite-service-2 kite-service-3`. Delete them once the new setup has run for a while.

## Moving the docs and web app off Vercel

1. Run `./deploy/docs.sh`, add docs.kite.onl as a public hostname on the Cloudflare tunnel pointing at nginx like api.kite.onl, and replace its DNS record (pointing at Vercel).
2. Once that works, do the same for kite.onl with `./deploy/web.sh`. Add www.kite.onl to the tunnel too, nginx redirects it to kite.onl.
