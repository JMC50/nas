# syntax=docker/dockerfile:1.6

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
# Phase 3+ will add: ffmpeg libreoffice
WORKDIR /app
COPY --from=build /out/server /app/server

ENV NODE_ENV=production \
    PORT=7777 \
    HOST=0.0.0.0 \
    NAS_DATA_DIR=/data/nas \
    NAS_ADMIN_DATA_DIR=/data/nas-admin \
    DB_PATH=/data/db/nas.sqlite \
    NAS_TEMP_DIR=/tmp/nas

EXPOSE 7777

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:7777/healthz || exit 1

CMD ["/app/server"]
