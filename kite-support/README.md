# kite-support

Discord support bot for [Kite](https://kite.onl). It answers questions about Kite with the `/ask` slash command, using the Kite documentation, every block of the flow editor and the current plans.

## How it works

1. `kite-support index` reads `kite-docs/docs` and the block catalog (`kite-service/pkg/flow/catalog.json`) and writes them as one markdown file, `internal/embedded/assets/knowledge.md`. Block pages in the docs only embed a component that shows the block's settings, so the settings from the catalog take its place. The file isn't committed, `deploy/support.sh` rebuilds it on every deploy so the bot always answers from the current docs.
2. `kite-support bot` connects to Discord and registers `/ask <question>`. Every question is sent to the model with all of the knowledge, about 30k tokens, which the provider caches. The plans are fetched from the Kite API every hour and added after it. Follow-ups asked from an answer include the questions and answers before them.

## Configuration

Copy `kite-support.example.toml` to `kite-support.toml` and fill in the Discord and OpenAI credentials. Every key can also be set via environment variables: `KITE_SUPPORT_<SECTION>__<KEY>` (double underscore between section and key). For example:

```
KITE_SUPPORT_DISCORD__TOKEN=...
KITE_SUPPORT_DISCORD__APP_ID=...
KITE_SUPPORT_OPENAI__API_KEY=...
```

## Local development

```
# build the knowledge (writes internal/embedded/assets/knowledge.md)
go run . index

# run the bot — `go run` recompiles, picking up the freshly written knowledge
go run . bot
```

The bot reads its knowledge from a file embedded in the binary at compile time. After running `index`, the next `go build` (or `go run`) embeds the new data, so a single binary is all that ships. A binary built without running `index` first refuses to start. The `/ask` command is registered globally on first launch. Global slash commands can take up to an hour to propagate.

## Docker

The supplied `Dockerfile` runs `index` against the in-tree docs, then builds the binary with the knowledge embedded.

```
docker build -f kite-support/Dockerfile -t kite-support .

docker run -d \
  -e KITE_SUPPORT_DISCORD__TOKEN=... \
  -e KITE_SUPPORT_DISCORD__APP_ID=... \
  -e KITE_SUPPORT_OPENAI__API_KEY=... \
  kite-support
```
