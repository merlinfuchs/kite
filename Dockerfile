FROM golang:latest as builder
WORKDIR /root/
COPY . .

# Install NodeJS (https://github.com/nodesource/distributions#installation-instructions)
RUN apt-get update
RUN apt-get install -y ca-certificates curl gnupg build-essential
RUN mkdir -p /etc/apt/keyrings
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg

RUN echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_18.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list

RUN apt-get update
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
ENV KITE_APP__PUBLIC_BASE_URL=http://localhost:8080
CMD ./kite-service database migrate postgres up; ./kite-service server start
