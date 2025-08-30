#!/bin/bash
# Development environment setup script
# This script helps set up the complete development environment

set -e

echo "🚀 Radarr Go Development Environment Setup"
echo "=========================================="

# Check if running on supported OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
else
    echo "❌ Unsupported OS: $OSTYPE"
    exit 1
fi

echo "📋 Detected OS: $OS"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install tools on macOS
install_macos() {
    echo "🍺 Installing tools via Homebrew..."
    
    if ! command_exists brew; then
        echo "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    
    # Install required tools
    if ! command_exists go; then
        echo "Installing Go..."
        brew install go
    fi
    
    if ! command_exists docker; then
        echo "Installing Docker..."
        brew install docker
    fi
    
    if ! command_exists docker-compose; then
        echo "Installing Docker Compose..."
        brew install docker-compose
    fi
    
    if ! command_exists node; then
        echo "Installing Node.js..."
        brew install node
    fi
    
    if ! command_exists make; then
        echo "Installing Make..."
        brew install make
    fi
}

# Function to install tools on Linux
install_linux() {
    echo "🐧 Installing tools for Linux..."
    
    # Update package list
    sudo apt update
    
    # Install Go
    if ! command_exists go; then
        echo "Installing Go..."
        sudo apt install -y golang-go
    fi
    
    # Install Docker
    if ! command_exists docker; then
        echo "Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER
        rm get-docker.sh
        echo "⚠️  Please log out and back in for Docker group changes to take effect"
    fi
    
    # Install Docker Compose
    if ! command_exists docker-compose; then
        echo "Installing Docker Compose..."
        sudo apt install -y docker-compose
    fi
    
    # Install Node.js
    if ! command_exists node; then
        echo "Installing Node.js..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
    
    # Install Make
    if ! command_exists make; then
        echo "Installing Make..."
        sudo apt install -y make
    fi
    
    # Install curl if not present
    if ! command_exists curl; then
        echo "Installing curl..."
        sudo apt install -y curl
    fi
}

# Install tools based on OS
if [[ "$OS" == "macos" ]]; then
    install_macos
elif [[ "$OS" == "linux" ]]; then
    install_linux
fi

echo "✅ System tools installation complete"

# Set up Go development tools
echo "🔧 Setting up Go development tools..."
make setup-backend

# Check environment
echo "🔍 Checking development environment..."
make check-env

# Create development directories
echo "📁 Creating development directories..."
mkdir -p data movies web/static web/templates tmp
mkdir -p web/frontend/src/components web/frontend/src/pages web/frontend/src/hooks
mkdir -p web/frontend/src/services web/frontend/src/utils web/frontend/public

# Create basic config if it doesn't exist
if [[ ! -f "config.yaml" ]]; then
    echo "📝 Creating basic development configuration..."
    cat > config.yaml << EOF
# Development configuration for Radarr Go
server:
  port: 7878
  host: "0.0.0.0"
  url_base: ""
  enable_ssl: false

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  username: "radarr_dev"
  password: "dev_password"
  name: "radarr_dev"
  ssl_mode: "disable"
  max_open_connections: 25
  max_idle_connections: 25

log:
  level: "debug"
  format: "json"
  file: ""

auth:
  method: "basic"
  api_key: "dev-api-key-12345"

storage:
  data_directory: "./data"
  movies_directory: "./movies"
EOF
fi

# Initialize git pre-commit hooks if available
if command_exists pre-commit; then
    echo "🪝 Setting up pre-commit hooks..."
    pre-commit install
else
    echo "💡 Consider installing pre-commit for automatic quality checks:"
    echo "   pip install pre-commit && pre-commit install"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Start the development environment: make dev-full"
echo "2. Or start backend only: make dev"
echo "3. Or start databases only: make test-db-up"
echo "4. Check the development guide: cat DEVELOPMENT.md"
echo ""
echo "Available services (when running make dev-full):"
echo "- Backend API: http://localhost:7878"
echo "- Database Admin: http://localhost:8081"
echo "- Monitoring: http://localhost:9090 (Prometheus)"
echo "- Grafana: http://localhost:3001"
echo "- Tracing: http://localhost:16686 (Jaeger)"
echo ""
echo "Happy coding! 🚀"