FROM golang:1.27-alpine3.24 AS builder

ARG GIT_COMMIT=none
ARG BUILD_TIME=unknown
ARG TAG=latest

ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN LDFLAGS="-s -w -X 'main.tag=${TAG}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.commit=${GIT_COMMIT}'" \
    && go build -trimpath -ldflags="${LDFLAGS}" -o bot ./cmd/bot \
    && go build -trimpath -ldflags="${LDFLAGS}" -o cli ./cmd/cli

FROM alpine:3.24

LABEL org.opencontainers.image.source=https://github.com/SkinonikS/discord-bot
LABEL org.opencontainers.image.description="Discord Bot Image"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /app

COPY --from=builder /build/bot /build/cli ./

COPY config/ ./config/
COPY i18n/ ./i18n/
COPY migrations/ ./migrations/

ENV APP_ROOT_DIR=/app

ENTRYPOINT ["./bot"]
