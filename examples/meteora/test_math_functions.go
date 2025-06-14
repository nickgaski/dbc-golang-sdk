package main

import (
	"fmt"
	"math/big"

	"dbc-golang-sdk/pkg/dbc"
)

func main() {
	fmt.Println("🧮 Meteora DBC - Mathematical Functions Test")
	fmt.Println("===========================================")

	// Test 1: Basic Constants
	fmt.Println("\n1️⃣ Testing Constants:")
	fmt.Printf("   Fee Denominator: %d\n", dbc.FEE_DENOMINATOR)
	fmt.Printf("   Max Fee BPS: %d (%.2f%%)\n", dbc.MAX_FEE_BPS, float64(dbc.MAX_FEE_BPS)/100)
	fmt.Printf("   Min Fee BPS: %d (%.4f%%)\n", dbc.MIN_FEE_BPS, float64(dbc.MIN_FEE_BPS)/100)
	fmt.Printf("   Basis Point Max: %d\n", dbc.BASIS_POINT_MAX)
	fmt.Printf("   Max Curve Points: %d\n", dbc.MAX_CURVE_POINT)

	// Test 2: Big Number Constants
	fmt.Println("\n2️⃣ Testing Big Number Constants:")
	fmt.Printf("   U64 Max: %s\n", dbc.U64_MAX.String())
	fmt.Printf("   Min Sqrt Price: %s\n", dbc.MIN_SQRT_PRICE.String())
	fmt.Printf("   Max Sqrt Price: %s\n", dbc.MAX_SQRT_PRICE.String())
	fmt.Printf("   One Q64: %s\n", dbc.ONE_Q64.String())

	// Test 3: Price Calculations
	fmt.Println("\n3️⃣ Testing Price Calculations:")
	
	// Example: Convert sqrt price to normal price
	// sqrt_price in Q64.64 format represents sqrt(price)
	sqrtPrice := big.NewInt(4295048016) // Min sqrt price
	
	// To get actual price: (sqrt_price / 2^64)^2
	q64 := new(big.Int).Lsh(big.NewInt(1), 64)
	priceRatio := new(big.Float).Quo(
		new(big.Float).SetInt(sqrtPrice),
		new(big.Float).SetInt(q64),
	)
	
	price := new(big.Float).Mul(priceRatio, priceRatio)
	priceFloat, _ := price.Float64()
	
	fmt.Printf("   Example Sqrt Price: %s\n", sqrtPrice.String())
	fmt.Printf("   Actual Price: %f\n", priceFloat)

	// Test 4: Fee Calculations
	fmt.Println("\n4️⃣ Testing Fee Calculations:")
	
	swapAmount := uint64(1000000) // 1 USDC (6 decimals)
	feeBps := uint64(30)          // 0.3%
	
	feeAmount := (swapAmount * feeBps) / dbc.BASIS_POINT_MAX
	netAmount := swapAmount - feeAmount
	
	fmt.Printf("   Swap Amount: %d\n", swapAmount)
	fmt.Printf("   Fee BPS: %d (%.2f%%)\n", feeBps, float64(feeBps)/100)
	fmt.Printf("   Fee Amount: %d\n", feeAmount)
	fmt.Printf("   Net Amount: %d\n", netAmount)

	// Test 5: Bonding Curve Mathematics
	fmt.Println("\n5️⃣ Testing Bonding Curve Mathematics:")
	
	// Example linear bonding curve calculation
	// Price = base_price + (supply * slope)
	basePrice := 1.0
	supply := 1000000.0
	slope := 0.000001
	
	currentPrice := basePrice + (supply * slope)
	
	fmt.Printf("   Base Price: $%.6f\n", basePrice)
	fmt.Printf("   Current Supply: %.0f tokens\n", supply)
	fmt.Printf("   Slope: %.6f\n", slope)
	fmt.Printf("   Current Price: $%.6f\n", currentPrice)

	// Test 6: Liquidity Distribution
	fmt.Println("\n6️⃣ Testing Liquidity Distribution:")
	
	// Example curve points for liquidity distribution
	curvePoints := []struct {
		SqrtPrice string
		Liquidity string
	}{
		{"4295048016", "1000000000000000000"},
		{"6442572024", "2000000000000000000"},
		{"8590096032", "3000000000000000000"},
	}
	
	fmt.Println("   Example Curve Points:")
	for i, point := range curvePoints {
		fmt.Printf("     Point %d: sqrt_price=%s, liquidity=%s\n", i+1, point.SqrtPrice, point.Liquidity)
	}

	// Test 7: Price Impact Calculation
	fmt.Println("\n7️⃣ Testing Price Impact Calculation:")
	
	reserveX := 1000000.0 // Base token reserve
	reserveY := 2000000.0 // Quote token reserve
	tradeSize := 50000.0  // Trade amount
	
	// Simplified constant product formula: x * y = k
	k := reserveX * reserveY
	
	// After trade: new_reserve_x = reserve_x + trade_size
	newReserveX := reserveX + tradeSize
	newReserveY := k / newReserveX
	
	outputAmount := reserveY - newReserveY
	
	// Price before and after
	priceBefore := reserveY / reserveX
	priceAfter := newReserveY / newReserveX
	priceImpact := ((priceAfter - priceBefore) / priceBefore) * 100
	
	fmt.Printf("   Reserve X: %.0f\n", reserveX)
	fmt.Printf("   Reserve Y: %.0f\n", reserveY)
	fmt.Printf("   Trade Size: %.0f\n", tradeSize)
	fmt.Printf("   Output Amount: %.0f\n", outputAmount)
	fmt.Printf("   Price Before: %.6f\n", priceBefore)
	fmt.Printf("   Price After: %.6f\n", priceAfter)
	fmt.Printf("   Price Impact: %.2f%%\n", priceImpact)

	// Test 8: Dynamic Fee Calculation
	fmt.Println("\n8️⃣ Testing Dynamic Fee Calculation:")
	
	// Example dynamic fee based on volatility
	baseFeeBps := uint64(25)                    // 0.25% base fee
	volatilityAccumulator := uint64(1500)       // 15% volatility
	maxVolatility := uint64(2000)               // 20% max volatility
	maxFeeBps := uint64(100)                    // 1% max fee
	
	// Dynamic fee increases with volatility
	volatilityRatio := float64(volatilityAccumulator) / float64(maxVolatility)
	if volatilityRatio > 1.0 {
		volatilityRatio = 1.0
	}
	
	dynamicFeeBps := baseFeeBps + uint64(float64(maxFeeBps - baseFeeBps) * volatilityRatio)
	
	fmt.Printf("   Base Fee: %d BPS (%.2f%%)\n", baseFeeBps, float64(baseFeeBps)/100)
	fmt.Printf("   Volatility: %d BPS (%.2f%%)\n", volatilityAccumulator, float64(volatilityAccumulator)/100)
	fmt.Printf("   Max Volatility: %d BPS (%.2f%%)\n", maxVolatility, float64(maxVolatility)/100)
	fmt.Printf("   Dynamic Fee: %d BPS (%.2f%%)\n", dynamicFeeBps, float64(dynamicFeeBps)/100)

	// Test 9: Migration Thresholds
	fmt.Println("\n9️⃣ Testing Migration Thresholds:")
	
	preMigrationSupply := uint64(1000000000)  // 1B tokens
	postMigrationSupply := uint64(800000000)  // 800M tokens (20% burned)
	migrationThreshold := uint64(500000000)   // 500M threshold
	
	burnPercentage := float64(preMigrationSupply - postMigrationSupply) / float64(preMigrationSupply) * 100
	thresholdMet := postMigrationSupply >= migrationThreshold
	
	fmt.Printf("   Pre-migration Supply: %d\n", preMigrationSupply)
	fmt.Printf("   Post-migration Supply: %d\n", postMigrationSupply)
	fmt.Printf("   Migration Threshold: %d\n", migrationThreshold)
	fmt.Printf("   Burn Percentage: %.1f%%\n", burnPercentage)
	fmt.Printf("   Threshold Met: %v\n", thresholdMet)

	// Test 10: Rate Limiter
	fmt.Println("\n🔟 Testing Rate Limiter:")
	
	maxRatePerSlot := uint64(100000)  // 100k tokens per slot
	currentSlot := uint64(1000)
	lastUpdateSlot := uint64(995)
	slotsElapsed := currentSlot - lastUpdateSlot
	
	// Refill rate limiter bucket
	refillAmount := slotsElapsed * maxRatePerSlot
	currentBucket := uint64(50000) // Current tokens in bucket
	maxBucket := maxRatePerSlot * 10 // 10 slots worth
	
	newBucket := currentBucket + refillAmount
	if newBucket > maxBucket {
		newBucket = maxBucket
	}
	
	fmt.Printf("   Max Rate Per Slot: %d\n", maxRatePerSlot)
	fmt.Printf("   Slots Elapsed: %d\n", slotsElapsed)
	fmt.Printf("   Refill Amount: %d\n", refillAmount)
	fmt.Printf("   Bucket Before: %d\n", currentBucket)
	fmt.Printf("   Bucket After: %d\n", newBucket)
	fmt.Printf("   Bucket Capacity: %d\n", maxBucket)

	fmt.Println("\n✅ All mathematical function tests completed!")
	fmt.Println("\n💡 These functions can be used to:")
	fmt.Println("   • Calculate swap quotes")
	fmt.Println("   • Determine optimal trade sizes")
	fmt.Println("   • Estimate price impact")
	fmt.Println("   • Design bonding curves")
	fmt.Println("   • Implement dynamic fees")
	fmt.Println("   • Handle migration logic")
}