FROM golang:latest AS builder
WORKDIR /root/
COPY . .

# Install NodeJS (https://github.com/nodesource/distributions#installation-instructions)
RUN apt-get update
RUN apt-get install -y ca-certificates curl gnupg build-essential
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get -y install nodejs

# Build website
ENV OUTPUT=export
ENV NEXT_PUBLIC_API_PUBLIC_BASE_URL=""
RUN cd kite-web && npm install && npm run build && cd ..

# Build backend
RUN cd kite-service && go build --tags "embedweb" && cd ..

FROM debian:stable-slim
WORKDIR /root/
COPY --from=builder /root/kite-service/kite-service .

RUN apt-get update
RUN apt-get install -y ca-certificates gnupg build-essential

EXPOSE 8080
CMD ./kite-service database migrate postgres up; ./kite-service server start
