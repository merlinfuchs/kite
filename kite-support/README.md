# kite-support

Discord support bot for [Kite](https://kite.onl). Indexes the Kite documentation and a build-time codebase summary, answers user questions in plain language via the `/ask` slash command.

The whole knowledge base is baked into a vector index at build time, so the running service has no dependency on the source tree.

## How it works

1. `kite-support summarize` — one-shot, run when the codebase changes meaningfully. Asks GPT-4 to read the `kite-service` source and emit `concepts.md` (user-facing concepts, no code). Commit the result.
2. `kite-support index` — walks `kite-docs/docs/` and `concepts.md`, chunks them, embeds with OpenAI `text-embedding-3-small`, writes `internal/embedded/assets/index.gob`. Commit the result.
3. `kite-support bot` — connects to Discord, registers `/ask <question>`, retrieves the top matching chunks for each question and asks GPT-4o-mini to phrase a friendly, code-free answer.

## Configuration

Copy `kite-support.example.toml` to `kite-support.toml` and fill in the Discord and OpenAI credentials. Every key can also be set via environment variables: `KITE_SUPPORT_<SECTION>__<KEY>` (double underscore between section and key). For example:

```
KITE_SUPPORT_DISCORD__TOKEN=...
KITE_SUPPORT_DISCORD__APP_ID=...
KITE_SUPPORT_OPENAI__API_KEY=...
```

## Local development

```
# regenerate the concept summary (writes concepts.md)
go run . summarize

# build the vector index (writes internal/embedded/assets/index.gob)
go run . index

# run the bot — `go run` recompiles, picking up the freshly written index
go run . bot
```

The bot reads its index from a blob embedded in the binary at compile time
(`internal/embedded/assets/index.gob`). After running `index`, the next
`go build` (or `go run`) embeds the new data, so a single binary is all
that ships. The `/ask` command is registered globally on first launch.
Global slash commands can take up to an hour to propagate.

## Docker

The supplied `Dockerfile` builds an indexer binary, runs `index` against the in-tree docs, then rebuilds so the final binary embeds the populated index.

```
docker build \
  --build-arg OPENAI_API_KEY=sk-... \
  -f kite-support/Dockerfile \
  -t kite-support .

docker run -d \
  -e KITE_SUPPORT_DISCORD__TOKEN=... \
  -e KITE_SUPPORT_DISCORD__APP_ID=... \
  -e KITE_SUPPORT_OPENAI__API_KEY=... \
  kite-support
```

The runtime container needs `KITE_SUPPORT_OPENAI__API_KEY` for query-time embedding and chat completion. No index file is shipped — it lives inside the binary.
