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
