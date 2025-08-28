#!/bin/bash

# NAS Local Build Script
# For development and local testing

set -e

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🐋 NAS Local Build Script${NC}"
echo "========================"
echo ""

cd "$(dirname "$0")/.."

# Check if .env exists, create if not
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}📝 Creating .env file from template...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created. Please review and update it.${NC}"
fi

# Load environment variables
source .env 2>/dev/null || true

# Set defaults
DATA_PATH=${DATA_PATH:-./data}
PORT=${PORT:-7777}

echo -e "${GREEN}📊 Configuration:${NC}"
echo "  Data Path: $DATA_PATH"
echo "  Port: $PORT"
echo ""

# Create directories
echo -e "${YELLOW}📁 Creating data directories...${NC}"
mkdir -p "${DATA_PATH}/files"
mkdir -p "${DATA_PATH}/admin" 
mkdir -p "${DATA_PATH}/database"
mkdir -p "${DATA_PATH}/temp"
mkdir -p "./config"

echo -e "${GREEN}✅ Directories created${NC}"

# Check build mode
echo "Build options:"
echo "1) Pull from registry (default)"
echo "2) Build locally from source"
echo ""
read -p "Select option (1-2) [1]: " build_option
build_option=${build_option:-1}

echo -e "${YELLOW}🛑 Stopping existing containers...${NC}"
docker-compose down 2>/dev/null || true

case $build_option in
    1)
        echo -e "${YELLOW}📥 Pulling latest image from registry...${NC}"
        docker-compose pull
        ;;
    2)
        echo -e "${YELLOW}🔨 Building Docker image locally...${NC}"
        # Temporarily modify compose file for local build
        cp docker-compose.yml docker-compose.yml.backup
        sed -i 's|image: ghcr.io.*|build: .|g' docker-compose.yml
        docker-compose build --no-cache
        mv docker-compose.yml.backup docker-compose.yml
        ;;
    *)
        echo -e "${RED}❌ Invalid option${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}🚀 Starting NAS application...${NC}"
export DOCKER_BUILDKIT=1
docker-compose up -d

echo ""
echo -e "${GREEN}✅ Container status:${NC}"
docker-compose ps

echo ""
echo -e "${YELLOW}🏥 Waiting for health check...${NC}"
sleep 15

# Check if container is running
if docker-compose ps | grep -q "Up.*healthy\|Up.*starting"; then
    echo -e "${GREEN}✅ NAS application is running!${NC}"
    echo ""
    echo -e "${GREEN}🌐 Access URLs:${NC}"
    echo "  Local: http://localhost:${PORT}"
    echo "  Network: http://$(hostname -I | awk '{print $1}'):${PORT}"
    echo ""
    echo -e "${GREEN}📊 Management Commands:${NC}"
    echo "  View logs: docker-compose logs -f"
    echo "  Stop app: docker-compose down"
    echo "  Restart: docker-compose restart"
    echo ""
else
    echo -e "${RED}❌ Failed to start NAS application${NC}"
    echo ""
    echo -e "${YELLOW}📋 Checking logs:${NC}"
    docker-compose logs
    echo ""
    echo -e "${YELLOW}💡 Troubleshooting tips:${NC}"
    echo "  1. Check if port ${PORT} is already in use"
    echo "  2. Verify .env file configuration"
    echo "  3. Check data directory permissions"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Build completed successfully!${NC}"