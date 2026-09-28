# Deploying

Scripts to deploy kite.onl, docs.kite.onl and api.kite.onl from a local checkout to a single server. nginx serves the web app and docs as static files, so deploying them never restarts the bots.

```shell
cp deploy/.env.example deploy/.env   # once, then fill it in

./deploy/web.sh                # kite.onl
./deploy/docs.sh               # docs.kite.onl
./deploy/service.sh            # api.kite.onl and the bots
./deploy/service.sh --rollback # restart with the previous binary
```

They need `pnpm`, `go`, `rsync` and SSH access to the server.

## Server setup

1. Add the server blocks from `deploy/nginx.conf` next to the api.kite.onl config and run `nginx -t && systemctl reload nginx`.
2. Add kite.onl and docs.kite.onl as public hostnames on the Cloudflare tunnel, pointing at nginx like api.kite.onl. The existing DNS records for them (pointing at Vercel) have to be removed first.
3. Create `WEB_DIR` and `DOCS_DIR` and make them writable by the SSH user.
4. Put `kite.toml` with `cluster_count = 4` in `SERVICE_DIR` and run the clusters from a systemd template unit like this one, as `kite-service@0` to `kite-service@3`. Each cluster gets its index and API port from the instance name, and migrations run on every start.

   ```ini
   # /etc/systemd/system/kite-service@.service
   [Unit]
   Description=Kite cluster %i
   After=network-online.target postgresql.service
   Wants=network-online.target

   [Service]
   WorkingDirectory=/opt/kite
   Environment=KITE_CLUSTER_INDEX=%i
   Environment=KITE_API__PORT=808%i
   ExecStartPre=/opt/kite/kite-service database migrate postgres up
   ExecStart=/opt/kite/kite-service server start
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

5. If the SSH user isn't root, it needs write access to `SERVICE_DIR` and permission to run `sudo systemctl restart` for the units.

Clusters restart one after another, so for a moment some run the new binary and some the old one against the migrated database. Migrations need to stay compatible with the previous version.
