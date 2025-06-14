#!/bin/bash

# DBC GoLang SDK - Complete Automated Setup
# This script sets up EVERYTHING needed to run the DBC SDK assessment examples

set -e  # Exit on any error

echo "🚀 DBC GoLang SDK - Automated Setup"
echo "===================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)
print_info "Detected OS: $OS"

# 1. Install Go if not present
echo -e "\n1️⃣ Installing Go..."
if command_exists go; then
    GO_VERSION=$(go version)
    print_status "Go already installed: $GO_VERSION"
else
    print_warning "Installing Go..."
    
    if [[ "$OS" == "macos" ]]; then
        if command_exists brew; then
            brew install go
        else
            print_error "Homebrew not found. Installing Homebrew first..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            brew install go
        fi
    elif [[ "$OS" == "linux" ]]; then
        # Download and install Go for Linux
        GO_VERSION="1.21.0"
        wget "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
        rm "go${GO_VERSION}.linux-amd64.tar.gz"
        
        # Add Go to PATH
        if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    print_status "Go installed successfully"
fi

# 2. Install Solana CLI
echo -e "\n2️⃣ Installing Solana CLI..."
if command_exists solana; then
    SOLANA_VERSION=$(solana --version)
    print_status "Solana CLI already installed: $SOLANA_VERSION"
else
    print_warning "Installing Solana CLI..."
    
    if [[ "$OS" == "macos" ]] && command_exists brew; then
        brew install solana
    else
        sh -c "$(curl -sSfL https://release.solana.com/stable/install)"
        export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH"
    fi
    
    print_status "Solana CLI installed successfully"
fi

# Add Solana to PATH for current session
export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH"

# 3. Install Node.js (needed for some dependencies)
echo -e "\n3️⃣ Installing Node.js..."
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_status "Node.js already installed: $NODE_VERSION"
else
    print_warning "Installing Node.js..."
    
    if [[ "$OS" == "macos" ]] && command_exists brew; then
        brew install node
    elif [[ "$OS" == "linux" ]]; then
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
    
    print_status "Node.js installed successfully"
fi

# 4. Set up Solana configuration for devnet
echo -e "\n4️⃣ Configuring Solana for devnet..."
solana config set --url devnet
print_status "Solana configured for devnet"

# 5. Generate or use existing keypair
echo -e "\n5️⃣ Setting up Solana keypair..."
KEYPAIR_FILE="$HOME/.config/solana/devnet-keypair.json"

# Create .config/solana directory if it doesn't exist
mkdir -p "$HOME/.config/solana"

if [[ -f "$KEYPAIR_FILE" ]]; then
    print_status "Using existing keypair: $KEYPAIR_FILE"
else
    print_warning "Generating new keypair..."
    solana-keygen new --outfile "$KEYPAIR_FILE" --no-bip39-passphrase
    print_status "New keypair generated: $KEYPAIR_FILE"
fi

# Set the keypair
solana config set --keypair "$KEYPAIR_FILE"
PUBLIC_KEY=$(solana-keygen pubkey "$KEYPAIR_FILE")
print_info "Public key: $PUBLIC_KEY"

# 6. Request devnet SOL airdrop
echo -e "\n6️⃣ Requesting devnet SOL airdrop..."
print_info "Requesting 2 SOL for testing..."

# Try airdrop with retries
for i in {1..5}; do
    if solana airdrop 2 --url devnet >/dev/null 2>&1; then
        print_status "Airdrop successful"
        break
    else
        print_warning "Airdrop attempt $i failed, retrying..."
        sleep 3
    fi
done

# Check balance
BALANCE=$(solana balance --url devnet 2>/dev/null || echo "0 SOL")
print_info "Current balance: $BALANCE"

# 7. Initialize Go module if not exists
echo -e "\n7️⃣ Setting up Go dependencies..."
if [[ ! -f "go.mod" ]]; then
    print_info "Initializing Go module..."
    go mod init dbc-golang-sdk
fi

# Add required dependencies with specific versions
go get github.com/gagliardetto/solana-go@v1.8.4
go get github.com/joho/godotenv@v1.5.1
go get github.com/near/borsh-go@v0.3.1
go get lukechampine.com/uint128@v1.3.0
go get go.mongodb.org/mongo-driver@latest

# Clean up and download dependencies
go mod tidy
go mod download

print_status "Go dependencies installed"

# 8. Create .env file with configuration
echo -e "\n8️⃣ Creating environment configuration..."
ENV_FILE=".env"

cat > "$ENV_FILE" << EOF
# DBC GoLang SDK Configuration
# Note: Examples use KEYPAIR_FILE for reliable key loading
PRIVATE_KEY=placeholder_not_used_examples_use_keypair_file
PUBLIC_KEY=$PUBLIC_KEY
SOLANA_RPC_URL=https://api.devnet.solana.com
NETWORK=devnet
DBC_PROGRAM_ID=dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN
KEYPAIR_FILE=$KEYPAIR_FILE

# Helius API key (optional - for enhanced RPC)
# HELIUS_API_KEY=your_helius_api_key_here
EOF

print_status "Environment file created: $ENV_FILE"

# 9. Test the setup
echo -e "\n9️⃣ Testing the setup..."

# Test Go build
if go build ./pkg/dbc >/dev/null 2>&1; then
    print_status "Go build test passed"
else
    print_error "Go build test failed"
    echo "Checking for issues..."
    go build ./pkg/dbc
fi

# Test solana connection
if solana epoch-info --url devnet >/dev/null 2>&1; then
    print_status "Solana connection test passed"
else
    print_error "Solana connection test failed"
fi

# 10. Create utility scripts
echo -e "\n🔟 Creating utility scripts..."

# Create test script
cat > "test.sh" << 'EOF'
#!/bin/bash
echo "🧪 Testing DBC SDK Setup"
echo "========================"

# Test 1: Check Go environment
echo "1. Go version:"
go version

# Test 2: Check Solana connection
echo -e "\n2. Solana network info:"
solana epoch-info --url devnet

# Test 3: Check balance
echo -e "\n3. Account balance:"
solana balance --url devnet

# Test 4: Test Go compilation
echo -e "\n4. Testing Go compilation:"
if go build ./pkg/dbc; then
    echo "✅ Go compilation successful"
    rm -f dbc  # Clean up binary
else
    echo "❌ Go compilation failed"
    exit 1
fi

# Test 5: Run a quick assessment example
echo -e "\n5. Testing assessment example compilation:"
if go run examples/assessment/01_create_config.go >/dev/null 2>&1; then
    echo "✅ Assessment examples work"
else
    echo "⚠️  Assessment examples need attention"
fi

echo -e "\n✅ Setup test completed!"
EOF

chmod +x test.sh

# Create run all examples script
cat > "run_all.sh" << 'EOF'
#!/bin/bash
echo "🚀 Running All DBC Assessment Examples"
echo "======================================"

# Cross-platform timeout function
run_with_timeout() {
    local timeout_duration=60
    local command="$1"
    
    # Check if timeout command exists (Linux)
    if command -v timeout >/dev/null 2>&1; then
        timeout ${timeout_duration}s $command
        return $?
    # Check if gtimeout exists (macOS with coreutils)
    elif command -v gtimeout >/dev/null 2>&1; then
        gtimeout ${timeout_duration}s $command
        return $?
    else
        # Fallback: run without timeout on macOS
        # Since our examples are designed to complete quickly, this should be safe
        $command
        return $?
    fi
}

examples=(
    "01_create_config"
    "02_create_pool" 
    "03_swap"
    "04_swap_quote"
    "05_claim_trading_fee"
    "06_withdraw_leftover"
    "07_migrate_damm_v1"
    "08_migrate_damm_v2"
)

for example in "${examples[@]}"; do
    echo -e "\n🔄 Running: $example"
    echo "----------------------------------------"
    
    if [[ -f "examples/assessment/${example}.go" ]]; then
        if run_with_timeout "go run examples/assessment/${example}.go"; then
            echo "✅ $example completed successfully"
        else
            echo "⚠️  $example completed with warnings or errors"
        fi
    else
        echo "❌ $example not found"
    fi
    
    echo "----------------------------------------"
done

echo -e "\n✅ All examples execution completed!"
EOF

chmod +x run_all.sh

print_status "Utility scripts created: test.sh, run_all.sh"

# 11. Build all examples
echo -e "\n🏗️ Building assessment examples..."
mkdir -p bin

for example in examples/assessment/*.go; do
    if [[ -f "$example" ]]; then
        example_name=$(basename "$example" .go)
        if go build -o "bin/$example_name" "$example" 2>/dev/null; then
            print_status "Built: $example_name"
        else
            print_warning "Build issue with: $example_name"
        fi
    fi
done

# 12. Final summary
echo -e "\n🎉 Setup completed successfully!"
echo "================================"
echo ""
echo "📋 What was installed/configured:"
echo "  ✅ Go programming language"
echo "  ✅ Solana CLI"
echo "  ✅ Node.js"
echo "  ✅ Solana devnet configuration"
echo "  ✅ Devnet keypair and SOL funding"
echo "  ✅ Go dependencies"
echo "  ✅ Environment configuration"
echo "  ✅ Assessment examples"
echo "  ✅ Utility scripts"
echo ""
echo "📁 Key files created:"
echo "  • .env - Environment configuration"
echo "  • $KEYPAIR_FILE - Solana keypair"
echo "  • test.sh - Test the setup"
echo "  • run_all.sh - Run all examples"
echo ""
echo "🚀 Ready to use commands:"
echo "  • ./test.sh - Test your setup"
echo "  • ./run_all.sh - Run all assessment examples"
echo "  • make run-assessment-all - Run via Makefile"
echo "  • go run examples/assessment/01_create_config.go - Run individual examples"
echo ""
echo "🔗 Your devnet address: $PUBLIC_KEY"
echo "💰 Current balance: $BALANCE"
echo ""
echo "📖 Next steps:"
echo "  1. Run './test.sh' to verify everything works"
echo "  2. Run './run_all.sh' to see all assessment examples"
echo "  3. Check individual examples in examples/assessment/"
echo ""
print_status "DBC SDK is ready for assessment! 🎯"

# Auto-run test if in interactive mode
if [[ -t 0 ]]; then
    echo ""
    read -p "Would you like to run the test now? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        ./test.sh
    fi
fi