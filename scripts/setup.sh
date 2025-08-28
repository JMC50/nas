#!/bin/bash

# NAS Setup & Upgrade Script
# One-click installation and upgrade for Docker-based NAS

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAS_DIR="/opt/nas"
DATA_DIR="/mnt/nas-storage"
DEFAULT_PORT=7777

echo -e "${GREEN}🚀 NAS Installation & Upgrade Script${NC}"
echo "==================================="
echo ""

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if running with appropriate privileges
check_permissions() {
    if [ "$EUID" -eq 0 ]; then
        print_warn "Running as root. This is not recommended for normal operation."
        print_info "Script will create a 'nasuser' for running the service."
    else
        print_info "Running as user: $(whoami)"
        # Check if user can use sudo
        if ! sudo -n true 2>/dev/null; then
            print_error "This script requires sudo privileges"
            exit 1
        fi
    fi
}

# Install Docker and Docker Compose
install_docker() {
    if command -v docker >/dev/null 2>&1; then
        print_info "Docker is already installed"
        docker --version
    else
        print_step "Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER
        rm get-docker.sh
        print_info "Docker installed successfully"
    fi

    if command -v docker-compose >/dev/null 2>&1; then
        print_info "Docker Compose is already installed"
        docker-compose --version
    else
        print_step "Installing Docker Compose..."
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        print_info "Docker Compose installed successfully"
    fi
}

# Create directories and setup permissions
setup_directories() {
    print_step "Setting up directories..."
    
    sudo mkdir -p $NAS_DIR
    sudo mkdir -p $DATA_DIR/{files,admin,database,temp}
    
    # Create nasuser if doesn't exist
    if ! id "nasuser" &>/dev/null; then
        sudo useradd -r -s /bin/false nasuser
        print_info "Created system user: nasuser"
    fi
    
    # Set proper ownership
    sudo chown -R $USER:$USER $NAS_DIR
    sudo chown -R $USER:$USER $DATA_DIR
    
    print_info "Directories created: $NAS_DIR, $DATA_DIR"
}

# Download or update compose file
setup_compose() {
    print_step "Setting up Docker Compose configuration..."
    
    cd $NAS_DIR
    
    # If updating, backup existing config
    if [ -f "docker-compose.yml" ]; then
        cp docker-compose.yml docker-compose.yml.backup
        print_info "Backed up existing configuration"
    fi
    
    # Create docker-compose.yml
    cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # NAS Application
  nas:
    image: ghcr.io/your-org/nas:latest  # TODO: Update with actual registry
    container_name: nas-app
    restart: unless-stopped
    ports:
      - "${PORT:-7777}:7777"
    volumes:
      - ${DATA_PATH:-./data}:/app/data
      - ${CONFIG_PATH:-./config}:/app/config
    env_file:
      - .env
    environment:
      - NODE_ENV=production
      - DATA_PATH=/app/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7777/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    labels:
      # Watchtower auto-update
      - "com.centurylinklabs.watchtower.enable=true"
    networks:
      - nas-network

  # Auto-update agent
  watchtower:
    image: containrrr/watchtower:latest
    container_name: nas-watchtower
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WATCHTOWER_CLEANUP=true
      - WATCHTOWER_POLL_INTERVAL=300
      - WATCHTOWER_INCLUDE_RESTARTING=true
      - WATCHTOWER_ROLLING_RESTART=true
      - TZ=${TZ:-UTC}
    command: nas-app
    networks:
      - nas-network

networks:
  nas-network:
    driver: bridge
EOF

    print_info "Docker Compose configuration created"
}

# Create environment file
create_env() {
    print_step "Creating environment configuration..."
    
    if [ ! -f ".env" ]; then
        cat > .env << EOF
# === NAS Configuration ===
PORT=${DEFAULT_PORT}
DATA_PATH=${DATA_DIR}
CONFIG_PATH=${NAS_DIR}/config

# === Security (REQUIRED: Change these!) ===
JWT_SECRET=$(openssl rand -hex 32)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=changeme123

# === Auto-update Settings ===
WATCHTOWER_POLL_INTERVAL=300
TZ=$(timedatectl show -p Timezone --value 2>/dev/null || echo "UTC")

# === Optional Settings ===
NODE_ENV=production
DEBUG=false
MAX_UPLOAD_SIZE=5000
SESSION_TIMEOUT=60
EOF
        print_info "Environment file created: .env"
        print_warn "IMPORTANT: Edit .env file to set your admin password!"
    else
        print_info "Environment file already exists: .env"
    fi
}

# Deploy or update the application
deploy_nas() {
    print_step "Deploying NAS application..."
    
    cd $NAS_DIR
    
    # Pull latest images
    docker-compose pull
    
    # Start services
    docker-compose up -d
    
    print_info "NAS application deployed successfully"
}

# Display status and next steps
show_status() {
    echo ""
    echo -e "${GREEN}🎉 NAS Installation Complete!${NC}"
    echo "=============================="
    echo ""
    
    print_info "Service Status:"
    docker-compose ps
    
    echo ""
    print_info "Access Information:"
    echo "  🌐 Web Interface: http://localhost:$DEFAULT_PORT"
    echo "  📁 Data Directory: $DATA_DIR"
    echo "  ⚙️  Config Directory: $NAS_DIR"
    echo ""
    
    print_info "Next Steps:"
    echo "  1. Edit $NAS_DIR/.env to set your admin password"
    echo "  2. Access the web interface and complete setup"
    echo "  3. Configure your file storage settings"
    echo ""
    
    print_info "Management Commands:"
    echo "  📊 View logs: cd $NAS_DIR && docker-compose logs -f"
    echo "  🔄 Restart: cd $NAS_DIR && docker-compose restart"
    echo "  🛑 Stop: cd $NAS_DIR && docker-compose down"
    echo "  🔧 Update: $NAS_DIR/scripts/setup.sh --upgrade"
    echo ""
    
    print_warn "Important Notes:"
    echo "  - Auto-updates are enabled (checks every 5 minutes)"
    echo "  - Change the default admin password immediately"
    echo "  - Set up regular backups of your data"
}

# Handle command line arguments
handle_args() {
    case "${1:-}" in
        --upgrade)
            print_info "Upgrading existing installation..."
            cd $NAS_DIR
            docker-compose pull
            docker-compose up -d
            docker-compose ps
            print_info "Upgrade completed!"
            exit 0
            ;;
        --status)
            cd $NAS_DIR 2>/dev/null || { print_error "NAS not installed"; exit 1; }
            docker-compose ps
            exit 0
            ;;
        --logs)
            cd $NAS_DIR 2>/dev/null || { print_error "NAS not installed"; exit 1; }
            docker-compose logs -f
            exit 0
            ;;
        --help|-h)
            echo "Usage: $0 [OPTION]"
            echo ""
            echo "Options:"
            echo "  --upgrade    Update existing installation"
            echo "  --status     Show service status"
            echo "  --logs       Show service logs"
            echo "  --help       Show this help message"
            echo ""
            exit 0
            ;;
        "")
            # No arguments, proceed with installation
            ;;
        *)
            print_error "Unknown argument: $1"
            print_info "Use --help for available options"
            exit 1
            ;;
    esac
}

# Main execution
main() {
    handle_args "$@"
    
    print_info "Starting NAS setup process..."
    echo ""
    
    check_permissions
    install_docker
    setup_directories
    setup_compose
    create_env
    deploy_nas
    show_status
    
    print_info "Setup completed successfully! 🎉"
}

# Run main function with all arguments
main "$@"