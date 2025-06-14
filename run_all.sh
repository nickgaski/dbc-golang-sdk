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
