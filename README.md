# 🪁 Kite

[![Release](https://github.com/merlinfuchs/kite/actions/workflows/release.yaml/badge.svg)](https://github.com/merlinfuchs/kite/releases)
[![Docker image](https://github.com/merlinfuchs/kite/actions/workflows/docker-push.yaml/badge.svg)](https://hub.docker.com/r/merlintor/kite)

[![Release](https://img.shields.io/github/v/release/merlinfuchs/kite)](https://github.com/merlinfuchs/kite/releases/latest)
[![MIT License](https://img.shields.io/github/license/merlinfuchs/kite)](LICENSE)
[![Discord Server](https://img.shields.io/discord/730045476459642900)](https://discord.gg/rNd9jWHnXh)

Make your own Discord Bot with Kite for free without a single line of code. With support for slash commands, buttons, and more.

![Flow Example](./example-flow.png)

This project is very much work in progress and not ready to be used for anything serious. Only some parts of the Discord API are covered and more complex logic can not be implemented.

## Self Hosting

This describes the easiest way to self host an instance of Kite by using a single binary that contains both the backend and frontend.

You can find prebuilt binaries of the server with the frontend files included [here](https://github.com/merlinfuchs/kite/releases/latest).

To run Kite you will also need to run a [Postgres](https://www.postgresql.org/) server alongside it, so it's recommended to use `docker-compose` to simplify the process.

### Configure the server

To configure the server you can create a file called `kite.toml` with the following fields:

```toml
[discord]
client_id = "..." # Your Discord client ID used for Oauth2
client_secret = "..." # Your Discord client secret used for Oauth2
```

You can also set the config values using environment variables. For example `KITE_DISCORD__CLIENT_ID` will set the discord client id.

### Using Docker (docker-compose)

Install Docker and docker-compose and create a docker-compose.yaml file with the following contents:

```yaml
version: "3.8"

services:
  postgres:
    image: postgres
    restart: always
    volumes:
      - kite-local-postgres:/var/lib/postgresql/data
    environment:
      POSTGRES_USER: postgres
      POSTGRES_DB: kite
      PGUSER: postgres
      PGDATA: /var/lib/postgresql/data/pgdata
      POSTGRES_HOST_AUTH_METHOD: trust
    healthcheck:
      test: ["CMD", "pg_isready"]
      interval: 3s
      timeout: 30s
      retries: 3

  kite:
    image: merlintor/kite:latest
    restart: always
    ports:
      - "8080:8080"
    environment:
      - KITE_API__HOST=0.0.0.0
      - KITE_DATABASE__POSTGRES__HOST=postgres
      - KITE_DATABASE__POSTGRES__USER=postgres
      - KITE_DATABASE__POSTGRES__DB_NAME=kite
      - KITE_APP__PUBLIC_BASE_URL=http://localhost:8080
      - KITE_API__PUBLIC_BASE_URL=http://localhost:8080
    volumes:
      - ./kite.toml:/root/kite.toml
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  kite-local-postgres:
```

Run the file using `docker-compose up`. It will automatically mount the `kite.toml` file into the container. You should not configure postgres in your config file as it's using the postgres instance from the container.

Kite should now be accessible in your browser at [localhost:8080](http://localhost:8080).

### Building from source

#### Building the website

You can download NodeJS and NPM from [nodejs.org](https://nodejs.org/en/download/).

```shell
# Switch to the kite-web directory
cd kite-web

# Install dependencies
npm install

# Start the development server (optional)
npm run dev

# Build for embedded use in kite-service (recommended)
OUTPUT=export npm run build

# Build for standalone use
npm run build
```

#### Building the docs

You can download NodeJS and NPM from [nodejs.org](https://nodejs.org/en/download/).

```shell
# Switch to the kite-docs directory
cd kite-docs

# Install dependencies
npm install

# Start the development server (optional)
npm run start

# Build for production use
npm run build
```

#### Building the server (kite-service)

Install Go >=1.22 from [go.dev](https://go.dev/doc/install).

```shell
# Switch to the backend directory
cd kite-service
# or if you are in the kite-web / kite-docs directoy
cd ../kite-service

# Configure the server (see steps above)

# Run database migrations
go run main.go database migrate postgres up

# Start the development server (optional)
go run main.go server

# Build and include the kite-web files in the backend binary (build website first)
go build --tags  "embedweb"

# Build without including the frontend files in the backend binary (you need to serve them yourself)
go build
```

## High Level Progress

- [x] Slash Commands
  - [ ] Sub Commands
  - [x] Basic Placeholders
  - [ ] Advanced Placeholders
- [x] Message Templates (MVP)
  - [x] Embeds
  - [ ] Attachments
  - [ ] Interactive Components
  - [ ] Basic placeholders
  - [ ] Advanced Placeholders
- [ ] Event Listeners
- [ ] Stored Variables (WIP)
  - [x] Basic infrastrcuture
  - [ ] Connect variables to commands
