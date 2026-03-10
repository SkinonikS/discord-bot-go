FROM golang:1.26-trixie AS builder

ARG GIT_COMMIT=none
ARG BUILD_TIME=unknown
ARG TAG=latest
ARG LIBDAVE_VERSION=v1.1.1

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ libopus-dev pkg-config git curl unzip cmake make ca-certificates nasm

RUN git clone https://github.com/disgoorg/godave /godave \
    && cd /godave/scripts \
    && SHELL=/bin/sh ./libdave_install.sh ${LIBDAVE_VERSION}

ENV PKG_CONFIG_PATH=/root/.local/lib/pkgconfig

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X 'main.tag=${TAG}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.commit=${GIT_COMMIT}'" \
    -o discord-bot ./cmd/bot

FROM ubuntu:24.04

LABEL org.opencontainers.image.source=https://github.com/SkinonikS/discord-bot
LABEL org.opencontainers.image.description="Discord Bot Image"
LABEL org.opencontainers.image.licenses=MIT

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libopus0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/discord-bot .
COPY --from=builder /root/.local/lib/libdave.so /usr/local/lib/
RUN ldconfig

COPY config/ ./config/
COPY i18n/ ./i18n/
COPY migrations/ ./migrations/

ENV APP_ROOT_DIR=/app
ENV APP_ENV=production

ENTRYPOINT ["./discord-bot"]