# Run a private Kite instance

This setup runs Kite's editor, API, and bot engine as a private instance. Its
Discord account allowlist controls who can sign in. Friends also need to be
invited as collaborators on each bot they should edit.

Self-hosting is not officially supported by the Kite project. Keep backups of
the Postgres database and any configured object storage.

## Test it locally

1. Create a Discord application in the [Discord Developer Portal](https://discord.com/developers/applications).
2. In OAuth2, add `http://localhost:8080/v1/auth/login/callback` as a redirect URL.
3. Copy `self-hosting/.env.example` to `self-hosting/.env`, add the application's
   client ID and secret, generate an encryption key with
   `openssl rand -hex 32`, and add your Discord user ID to
   `KITE_SELF_HOSTED__ALLOWED_DISCORD_IDS`. Add a friend's ID there too if they
   should be able to sign in.
4. Start Kite from the repository root:

   ```sh
   docker compose -f self-hosting/docker-compose.yaml up --build
   ```

5. Open [http://localhost:8080](http://localhost:8080) and sign in with an
   allowlisted Discord account. Other Discord accounts are rejected. Invite a
   friend to a bot from that bot's collaborator settings after they sign in.

The first build may take a while. This compose file starts Postgres and Kite;
it omits object storage, so asset uploads are unavailable until you configure
S3-compatible storage.

## Deploy it on Railway

Railway can build the root `Dockerfile` directly from this repository. Create a
Postgres service and a Kite service in one Railway project, then set the Kite
service variables below. Configure a public domain for the Kite service and
use that same HTTPS URL for both public base URLs, for example
`https://my-kite.up.railway.app`.

```text
KITE_API__HOST=0.0.0.0
KITE_API__PORT=8080
KITE_API__PUBLIC_BASE_URL=https://YOUR-DOMAIN
KITE_API__SECURE_COOKIES=true
KITE_APP__PUBLIC_BASE_URL=https://YOUR-DOMAIN
KITE_DATABASE__POSTGRES__HOST=${{Postgres.PGHOST}}
KITE_DATABASE__POSTGRES__PORT=${{Postgres.PGPORT}}
KITE_DATABASE__POSTGRES__USER=${{Postgres.PGUSER}}
KITE_DATABASE__POSTGRES__PASSWORD=${{Postgres.PGPASSWORD}}
KITE_DATABASE__POSTGRES__DB_NAME=${{Postgres.PGDATABASE}}
KITE_DISCORD__CLIENT_ID=YOUR-DISCORD-CLIENT-ID
KITE_DISCORD__CLIENT_SECRET=YOUR-DISCORD-CLIENT-SECRET
KITE_ENCRYPTION__TOKEN_ENCRYPTION_KEY=YOUR-32-BYTE-HEX-KEY
KITE_SELF_HOSTED__ENABLED=true
KITE_SELF_HOSTED__ALLOWED_DISCORD_IDS=YOUR-ID,FRIEND-ID
```

In the Discord application's OAuth2 settings, set the redirect URL to
`https://YOUR-DOMAIN/v1/auth/login/callback`. Generate a persistent encryption
key once and keep it safe: changing it later prevents Kite from reading stored
Discord tokens. Railway's Postgres service must keep its persistent volume.

Uploads need S3-compatible object storage; without it, the editor and bots can
still be tried, but uploaded assets are disabled. AI features need your own
OpenAI or OpenRouter API key, and the provider may charge for usage.

Railway Free has small per-service resource and volume limits and only a small
monthly usage credit. A bot engine needs to stay online, so treat Free as a
smoke-test environment; monitor Railway usage and expect to move to a paid plan
or another always-on host if it runs out of resources. See Railway's current
[plan limits](https://docs.railway.com/pricing/plans) and
[usage billing](https://docs.railway.com/pricing/understanding-your-bill).

## What the self-hosted profile changes

When `self_hosted.enabled` is true, the configured default plan enables its
feature flags and receives high quotas, including collaborator access for up
to 10 users per bot. The profile also removes the hosted plan's five-minute
minimum for scheduled tasks. These are generous software quotas, not unlimited
compute: Discord, your host, object storage, and external API providers still
have their own limits or costs.

At least one Discord user ID is required in the allowlist when this profile is
enabled. For TOML configuration use an array:

```toml
[self_hosted]
enabled = true
allowed_discord_ids = ["YOUR_ID", "FRIEND_ID"]
```
