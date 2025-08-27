# NAS File Manager - Ubuntu-based Docker image
FROM ubuntu:22.04 AS base
WORKDIR /app

# Install Node.js 20 and system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    python3 \
    python3-pip \
    build-essential \
    sqlite3 \
    wget \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Development stage
FROM base AS development
ENV NODE_ENV=development
ENV PORT=7777
ENV HOST=0.0.0.0

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Build backend
WORKDIR /app/backend
RUN npm run build 2>/dev/null || npx tsc

# Build frontend
WORKDIR /app/frontend
RUN npm run build

WORKDIR /app

# Create data directories
RUN mkdir -p /app/data /app/admin-data /app/backend/db

# Expose ports
EXPOSE 7777 5050

# Development command
CMD ["npm", "run", "start"]

# Production stage
FROM ubuntu:22.04 AS production
WORKDIR /app

# Install only runtime dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    sqlite3 \
    wget \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create application user
RUN groupadd -r nasapp && useradd -r -g nasapp -s /bin/false nasapp

# Copy built application from development stage
COPY --from=development --chown=nasapp:nasapp /app/backend/dist ./backend/dist
COPY --from=development --chown=nasapp:nasapp /app/frontend/dist ./frontend/dist
COPY --from=development --chown=nasapp:nasapp /app/backend/src/entity ./backend/src/entity
COPY --from=development --chown=nasapp:nasapp /app/backend/src/migrations ./backend/src/migrations

# Copy production package files and install production dependencies
COPY --chown=nasapp:nasapp package*.json ./
RUN npm ci --only=production

# Create data directories with proper permissions
RUN mkdir -p /app/data /app/admin-data /app/db /tmp/nas && \
    chown -R nasapp:nasapp /app /tmp/nas

# Switch to application user
USER nasapp

# Environment variables for production
ENV NODE_ENV=production
ENV PORT=7777
ENV HOST=0.0.0.0
ENV NAS_DATA_DIR=/app/data
ENV NAS_ADMIN_DATA_DIR=/app/admin-data
ENV DB_PATH=/app/db
ENV NAS_TEMP_DIR=/tmp/nas

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:7777/ || exit 1

# Expose port
EXPOSE 7777

# Production command
CMD ["node", "backend/dist/index.js"]