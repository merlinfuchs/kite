# Deploying

Scripts to deploy kite.onl, docs.kite.onl and api.kite.onl from a local checkout to a single server. nginx serves the web app and docs as static files, so deploying them never restarts the bots, and routes API requests to the service cluster that runs the app.

```shell
cp deploy/.env.example deploy/.env   # once, then set DEPLOY_HOST

./deploy/web.sh                # kite.onl
./deploy/docs.sh               # docs.kite.onl
./deploy/service.sh            # api.kite.onl and the bots, one cluster after another
./deploy/service.sh --rollback # restart with the previous binary
./deploy/server.sh             # nginx config and systemd unit
```

They need `pnpm`, `go`, `rsync` and SSH access to the server.

## Configuration

`config.env` holds the production setup: number of clusters, ports, domains and paths. `nginx.conf` and `kite-service@.service` are generated from it by `generate.sh`. Don't edit them by hand, CI checks they're up to date.

To change the number of clusters, set `CLUSTER_COUNT`, run `./deploy/generate.sh`, commit, and run `./deploy/server.sh`. It restarts all clusters at once, since apps move between them, and switches nginx over to the new routing.

## Server setup

1. Put `kite.toml` in `SERVICE_DIR`. The cluster count, index and API port come from the systemd unit, so leave them out.
2. Remove the old api.kite.onl nginx config and the old kite-service unit, then run `./deploy/server.sh`. It installs the nginx config at `NGINX_CONF_PATH` and the unit, and starts the clusters as `kite-service@0` to `kite-service@N`.
3. Add kite.onl and docs.kite.onl as public hostnames on the Cloudflare tunnel, pointing at nginx like api.kite.onl. The existing DNS records for them (pointing at Vercel) have to be removed first.
4. Create `WEB_DIR` and `DOCS_DIR` and make them writable by the SSH user.
5. If the SSH user isn't root, it needs write access to `SERVICE_DIR` and passwordless or interactive `sudo`.

Clusters restart one after another on deploys, so for a moment some run the new binary and some the old one against the migrated database. Migrations need to stay compatible with the previous version.
