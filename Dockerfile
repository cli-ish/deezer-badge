FROM golang:1.20.3-alpine as builder
ENV USER=appuser
ENV UID=10001
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
RUN mkdir -p /srv/deezer-badge /srv/deezer-badge/release
WORKDIR /srv/deezer-badge
COPY ./internal internal/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
RUN go build -o /srv/deezer-badge/release/deezer-badge ./internal

FROM alpine:3.17
COPY --from=builder /srv/deezer-badge/release /srv/deezer-badge

ENV REDIS_HOST="redis-storage:6379"
ENV REDIS_PASS="password123456"
ENV APP_ID="123456"
ENV APP_SECRET="ffffffffffffffffffffffffffffffff"
ENV RETURN_URL="https://example.com/auth"

ENTRYPOINT ["/srv/deezer-badge/deezer-badge"]