# NAS File Manager - Optimized Alpine-based Docker image
FROM node:20-alpine AS builder
WORKDIR /app

# Install build dependencies for native modules
RUN apk add --no-cache python3 make g++ sqlite-dev

# Layer caching: Copy package files first
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY backend ./backend
COPY frontend ./frontend

# Build backend
WORKDIR /app/backend
RUN npm run build 2>/dev/null || npx tsc

# Build frontend  
WORKDIR /app/frontend
RUN npm run build

# Production stage
FROM node:20-alpine AS production
WORKDIR /app

# Install runtime essentials and create user
RUN apk add --no-cache sqlite curl && \
    addgroup -g 1001 -S nasapp && \
    adduser -S nasapp -u 1001 -G nasapp

# Copy built application from builder
COPY --from=builder --chown=nasapp:nasapp /app/backend/dist ./backend/dist
COPY --from=builder --chown=nasapp:nasapp /app/frontend/dist ./frontend/dist
COPY --from=builder --chown=nasapp:nasapp /app/backend/src/entity ./backend/src/entity
COPY --from=builder --chown=nasapp:nasapp /app/backend/src/migrations ./backend/src/migrations

# Copy production dependencies
COPY --chown=nasapp:nasapp package*.json ./
RUN npm ci --only=production --omit=dev && npm cache clean --force

# Create data directories
RUN mkdir -p /app/data /app/admin-data /app/db /tmp/nas && \
    chown -R nasapp:nasapp /app /tmp/nas

# Switch to non-root user
USER nasapp

# Environment variables with defaults
ENV NODE_ENV=production \
    PORT=7777 \
    HOST=0.0.0.0 \
    NAS_DATA_DIR=/app/data \
    NAS_ADMIN_DATA_DIR=/app/admin-data \
    DB_PATH=/app/db/nas.sqlite \
    NAS_TEMP_DIR=/tmp/nas

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:7777/health || exit 1

EXPOSE 7777

# Simple command without entrypoint
CMD ["node", "backend/dist/index.js"]