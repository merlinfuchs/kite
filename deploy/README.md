# Deploying

Scripts to deploy kite.onl, docs.kite.onl and api.kite.onl from a local checkout to a single server. Caddy serves the web app and docs as static files and proxies the API to kite-service, so deploying the web app or docs never restarts the bots.

```shell
cp deploy/.env.example deploy/.env   # once, then fill it in

./deploy/web.sh                # kite.onl
./deploy/docs.sh               # docs.kite.onl
./deploy/service.sh            # api.kite.onl and the bots
./deploy/service.sh --rollback # restart with the previous binary
```

They need `pnpm`, `go`, `rsync` and SSH access to the server.

## Server setup

1. Install Caddy and put `deploy/Caddyfile` at `/etc/caddy/Caddyfile`. If Cloudflare proxies the domains, set its SSL mode to Full (strict).
2. Create `WEB_DIR` and `DOCS_DIR` and make them writable by the SSH user.
3. Put `kite.toml` in `SERVICE_DIR` and run kite-service with a systemd unit like this one. It applies migrations on every start.

   ```ini
   [Unit]
   Description=Kite
   After=network-online.target postgresql.service
   Wants=network-online.target

   [Service]
   WorkingDirectory=/opt/kite
   ExecStartPre=/opt/kite/kite-service database migrate postgres up
   ExecStart=/opt/kite/kite-service server start
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

4. If the SSH user isn't root, it needs write access to `SERVICE_DIR` and permission to run `sudo systemctl restart <SERVICE_NAME>`.
