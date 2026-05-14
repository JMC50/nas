# syntax=docker/dockerfile:1.6

# --- Stage 1: build frontend (Vite + Svelte) ---
FROM node:22-alpine AS frontend-build
WORKDIR /work
COPY package*.json ./
RUN npm ci --prefer-offline --no-audit --silent
# Frontend has its own package.json (no workspaces at root) — install separately
# so deps declared only in frontend/package.json (e.g. pdfjs-dist) are present
# when vite resolves modules.
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci --prefer-offline --no-audit --silent
# VERSION at /work/VERSION — vite.config.ts reads `resolve(__dirname, "..", "VERSION")`
# which from /work/frontend/ resolves to /work/VERSION. Without this copy the
# readFileSync throws and __APP_VERSION__ falls back to "0.0.0-unknown".
COPY VERSION ./
COPY frontend ./frontend
RUN cd frontend && npx vite build
# SvelteKit (adapter-static) writes to /work/frontend/build

# --- Stage 2: build Go binary ---
FROM golang:1.25-alpine AS go-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- Stage 3: runtime ---
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget libreoffice ttf-dejavu fontconfig
RUN mkdir -p /tmp/soffice-home
WORKDIR /app
COPY --from=go-build /out/server /app/server
COPY --from=frontend-build /work/frontend/build /app/frontend/build

ENV NODE_ENV=production \
    PORT=7777 \
    HOST=0.0.0.0 \
    HOME=/tmp/soffice-home \
    NAS_DATA_DIR=/data/nas \
    NAS_ADMIN_DATA_DIR=/data/nas-admin \
    DB_PATH=/data/db/nas.sqlite \
    NAS_TEMP_DIR=/tmp/nas \
    FRONTEND_DIR=/app/frontend/build

EXPOSE 7777

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- "http://localhost:${PORT:-7777}/healthz" || exit 1

CMD ["/app/server"]
