---
sidebar_position: 6
---

# Self-hosting

:::warning

Self-hosting is not officially supported. These instructions are provided as-is for people who are comfortable running Docker, Postgres and their own Discord application. We don't answer self-hosting questions in issues or on Discord, and they may be closed without a reply. Bug reports that also affect the hosted version of Kite are welcome.

:::

Run Kite on your own machine with Docker Compose. This setup runs the Kite dashboard and API together, with PostgreSQL for the database and S3-compatible storage for uploaded files.

## Before you start

- Install Docker with the Compose plugin for the container setup.
- Have access to the Discord Developer Portal. Kite uses a Discord application to sign users into the dashboard.
- Install OpenSSL to generate the encryption key and service passwords.

The Compose example serves Kite on the Docker host at `http://localhost:8080`. PostgreSQL and object storage are accessible to Kite on the internal Docker network.

## Configure Discord login

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications).
2. Open **OAuth2** and add `http://localhost:8080/v1/auth/login/callback` under **Redirects**. Save your changes.
3. Copy the **Client ID** and **Client Secret** from the OAuth2 page.

## Configure Kite

In an empty directory, create `kite.toml`:

```toml
[discord]
client_id = "YOUR_DISCORD_CLIENT_ID"
client_secret = "YOUR_DISCORD_CLIENT_SECRET"

[encryption]
token_encryption_key = "YOUR_HEX_ENCRYPTION_KEY"
```

Run `openssl rand -hex 32` and paste the output into `token_encryption_key`. Keep a copy of this key with your backups. Kite needs the same key to read existing encrypted bot tokens.

Create a `.env` file in the same directory. Run `openssl rand -hex 32` separately for each password and replace the placeholders with the generated values.

```dotenv
POSTGRES_PASSWORD=YOUR_POSTGRES_PASSWORD
S3_PASSWORD=YOUR_STORAGE_PASSWORD
```

Keep `kite.toml` and `.env` private.

## Start Kite with Docker Compose

Save this as `compose.yaml` alongside `kite.toml` and `.env`:

```yaml
services:
  postgres:
    image: postgres:17
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_DB: kite
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in .env}"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres", "-d", "kite"]
      interval: 3s
      timeout: 5s
      retries: 10

  minio:
    image: pgsty/silo
    command: server --console-address ":9001" /data
    restart: unless-stopped
    environment:
      MINIO_ROOT_USER: kite
      MINIO_ROOT_PASSWORD: "${S3_PASSWORD:?Set S3_PASSWORD in .env}"
    volumes:
      - storage-data:/data

  kite:
    image: merlintor/kite:latest
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      KITE_API__HOST: "0.0.0.0"
      KITE_APP__PUBLIC_BASE_URL: "http://localhost:8080"
      KITE_API__PUBLIC_BASE_URL: "http://localhost:8080"
      KITE_DATABASE__POSTGRES__HOST: "postgres"
      KITE_DATABASE__POSTGRES__USER: "postgres"
      KITE_DATABASE__POSTGRES__DB_NAME: "kite"
      KITE_DATABASE__POSTGRES__PASSWORD: "${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in .env}"
      KITE_DATABASE__S3__ENDPOINT: "minio:9000"
      KITE_DATABASE__S3__ACCESS_KEY_ID: "kite"
      KITE_DATABASE__S3__SECRET_ACCESS_KEY: "${S3_PASSWORD:?Set S3_PASSWORD in .env}"
      KITE_DATABASE__S3__SECURE: "false"
    volumes:
      - ./kite.toml:/root/kite.toml:ro
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_started

volumes:
  postgres-data:
  storage-data:
```

From that directory, check the configuration and start the services:

```sh
docker compose config --quiet
docker compose up -d
docker compose ps
```

Open [http://localhost:8080](http://localhost:8080) on the machine running Docker and sign in with Discord. If Kite fails to start, run `docker compose logs kite` to check the startup logs.

To stop Kite, run `docker compose down`. PostgreSQL and uploaded files are stored in the named Docker volumes.

## Run a release binary

The [Kite releases page](https://github.com/merlinfuchs/kite/releases/latest) provides prebuilt binaries with the dashboard included. For a native installation, provide your own running PostgreSQL database and S3-compatible storage. The Compose example above keeps both services on its internal network.

Keep the Discord and encryption settings from [Configure Kite](#configure-kite), then add your database and storage settings to `kite.toml`. Replace these example values with the endpoints and credentials for your services.

```toml
[api]
public_base_url = "http://localhost:8080"

[app]
public_base_url = "http://localhost:8080"

[database.postgres]
host = "127.0.0.1"
port = 5432
user = "kite"
db_name = "kite"
password = "YOUR_POSTGRES_PASSWORD"

[database.s3]
endpoint = "127.0.0.1:9000"
access_key_id = "YOUR_STORAGE_ACCESS_KEY"
secret_access_key = "YOUR_STORAGE_SECRET_KEY"
secure = false
```

Set `secure = true` when the storage endpoint uses HTTPS. From the directory containing `kite.toml`, run the downloaded binary. These commands assume the executable is named `kite-service`.

```sh
./kite-service database migrate postgres up
./kite-service server start
```

Kite also accepts environment variables for configuration. For example, `KITE_DISCORD__CLIENT_ID` sets `[discord].client_id`.

## Build from source

Install Node.js, Corepack, and Go 1.26 or later. Run the frontend build from the repository root:

```sh
cd kite-web
corepack pnpm install --frozen-lockfile
OUTPUT=export NEXT_PUBLIC_API_PUBLIC_BASE_URL="" corepack pnpm run build
```

The export lets the Go server serve the dashboard and API from the same address. Build the combined binary next:

```sh
cd ../kite-service
go build -tags embedweb -o kite-service .
```

Follow [Run a release binary](#run-a-release-binary) to configure, migrate, and start your build. To build the dashboard and server separately, run `corepack pnpm run build` in `kite-web` without the export settings and run `go build` in `kite-service` without the embed tag. Serve the dashboard separately and set its API URL to match the server.

### Build the documentation

The docs have their own build. From the repository root, run:

```sh
cd kite-docs
corepack pnpm install --frozen-lockfile
corepack pnpm run build
```

For a local docs preview, run `corepack pnpm run start` and open [http://localhost:4000](http://localhost:4000).
