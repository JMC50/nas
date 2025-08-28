#!/bin/bash

set -e  

echo "🐋 NAS Docker Build & Deploy"
echo "========================="

cd "$(dirname "$0")/.."

if [ ! -f ".env" ]; then
    echo "❌  no .env file found"
    echo "Please create a .env file with DATA_PATH set."
    exit 1
fi

DATA_PATH=$(grep "^DATA_PATH=" .env | cut -d '=' -f2)
if [ -z "$DATA_PATH" ]; then
    echo "❌ .env file is missing DATA_PATH"
    exit 1
fi

echo "📁 Data Path: $DATA_PATH"

echo "📁 Creating necessary directories..."
sudo mkdir -p "${DATA_PATH}/files"
sudo mkdir -p "${DATA_PATH}/admin" 
sudo mkdir -p "${DATA_PATH}/database"
sudo mkdir -p "${DATA_PATH}/temp"
sudo chown -R $(id -u):$(id -g) "$DATA_PATH"

echo "🛑 Stopping existing containers..."
docker-compose down 2>/dev/null || true

echo "🔨 Building Docker image..."
docker-compose build --no-cache

echo "🚀 Starting NAS application..."
docker-compose up -d

echo "✅ Container status:"
docker-compose ps

echo "🏥 Waiting for health check..."
sleep 10

if docker-compose ps | grep -q "Up"; then
    echo "✅ NAS application is running!"
    echo "🌐 Access: http://localhost:8086"
    echo "📊 Logs: docker-compose logs -f"
else
    echo "❌ Failed to start NAS application"
    echo "📋 Check logs: docker-compose logs"
    exit 1
fi