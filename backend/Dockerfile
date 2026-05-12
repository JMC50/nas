# syntax=docker/dockerfile:1.6

# --- Stage 1: build frontend (Vite + Svelte) ---
FROM node:22-alpine AS frontend-build
WORKDIR /work
COPY package*.json ./
RUN npm ci --prefer-offline --no-audit --silent
COPY frontend ./frontend
WORKDIR /work/frontend
RUN npx vite build

# --- Stage 2: build Go binary ---
FROM golang:1.25-alpine AS go-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Stage 3: runtime ---
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
# Phase 3+ will add: ffmpeg libreoffice
WORKDIR /app
COPY --from=go-build /out/server /app/server
COPY --from=frontend-build /work/frontend/dist /app/frontend/dist

ENV NODE_ENV=production \
    PORT=7777 \
    HOST=0.0.0.0 \
    NAS_DATA_DIR=/data/nas \
    NAS_ADMIN_DATA_DIR=/data/nas-admin \
    DB_PATH=/data/db/nas.sqlite \
    NAS_TEMP_DIR=/tmp/nas \
    FRONTEND_DIR=/app/frontend/dist

EXPOSE 7777

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:7777/healthz || exit 1

CMD ["/app/server"]
