# Two-stage build: produce binary with empty embedded index, run index against
# OpenAI to populate the embed asset, then rebuild so the final binary embeds
# the populated index. The runtime image needs no source or asset files.
FROM golang:1.25-bookworm AS build
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.work go.work.sum ./
COPY kite-service ./kite-service
COPY kite-web ./kite-web
COPY kite-docs ./kite-docs
COPY kite-support ./kite-support
WORKDIR /src/kite-support

# Build a temporary indexer using the empty placeholder embed.
RUN go build -o /tmp/kite-support-indexer .

ARG OPENAI_API_KEY
RUN test -n "$OPENAI_API_KEY" || (echo "OPENAI_API_KEY build arg required" && exit 1)
RUN KITE_SUPPORT_OPENAI__API_KEY=$OPENAI_API_KEY /tmp/kite-support-indexer index

# Re-build to embed the populated index.
RUN go build -o /out/kite-support .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/kite-support /app/kite-support
CMD ["/app/kite-support", "bot"]
