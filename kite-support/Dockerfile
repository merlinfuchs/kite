# Builds the knowledge from the docs and the block catalog, then the binary
# with it embedded. The runtime image needs no source or asset files.
FROM golang:1.25-bookworm AS build
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.work go.work.sum ./
COPY kite-service ./kite-service
COPY kite-web ./kite-web
COPY kite-docs ./kite-docs
COPY kite-support ./kite-support
WORKDIR /src/kite-support

RUN go run . index
RUN go build -o /out/kite-support .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/kite-support /app/kite-support
CMD ["/app/kite-support", "bot"]
