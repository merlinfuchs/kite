FROM golang:1.26 AS builder
WORKDIR /root/
COPY . .

# Install NodeJS (https://github.com/nodesource/distributions#installation-instructions)
RUN apt-get update
# libopus-dev + pkg-config are needed to build the cgo Opus encoder used by
# the voice "Play Audio" block.
RUN apt-get install -y ca-certificates curl gnupg build-essential libopus-dev libopusfile-dev pkg-config
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get -y install nodejs

# Build website
ENV OUTPUT=export
ENV NEXT_PUBLIC_API_PUBLIC_BASE_URL=""
RUN corepack enable && cd kite-web && pnpm install --frozen-lockfile && pnpm run build && cd ..

# Build backend
RUN cd kite-service && go mod tidy && cd ..
RUN cd kite-service && CGO_ENABLED=1 go build --tags "embedweb" && cd ..

FROM debian:stable-slim
WORKDIR /root/
COPY --from=builder /root/kite-service/kite-service .

RUN apt-get update
# ffmpeg decodes uploaded audio files (mp3, etc.) for the voice "Play Audio"
# block; libopus0 is the runtime lib the cgo Opus encoder links against.
RUN apt-get install -y ca-certificates gnupg build-essential ffmpeg libopus0 libopusfile0

EXPOSE 8080
CMD ./kite-service database migrate postgres up; ./kite-service server start
