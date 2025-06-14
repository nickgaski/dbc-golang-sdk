package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dbc-golang-sdk/pkg/dbc"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

// Example script for SwapQuote as per assessment requirements
func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("📊 DBC Assessment - SwapQuote Example")
	fmt.Println("=====================================")

	// Get private key from keypair file
	keypairFile := os.Getenv("KEYPAIR_FILE")
	if keypairFile == "" {
		keypairFile = os.Getenv("HOME") + "/.config/solana/devnet-keypair.json"
	}

	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(keypairFile)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Create client
	config := dbc.DevnetConfig()
	client := dbc.NewDBCClientWithConfig(config, privateKey)

	ctx := context.Background()

	fmt.Printf("🔑 Using account: %s\n", client.Payer.PublicKey())

	// Check balance
	balance, err := client.RPCClient.GetBalance(ctx, client.Payer.PublicKey(), client.Commitment)
	if err != nil {
		log.Printf("Warning: Could not check balance: %v", err)
	} else {
		fmt.Printf("💰 Balance: %.4f SOL\n", float64(balance.Value)/1e9)
	}

	// Define quote parameters according to assessment requirements
	baseMint := dbc.GetNativeMint() // SOL
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)

	// Test different swap scenarios
	swapScenarios := []struct {
		name               string
		amount             uint64
		swapType           uint8
		slippage           float64
		includeFeesInQuote bool
	}{
		{"Small Buy", 1000000, 0, 0.01, true},        // 0.001 SOL buy with 1% slippage
		{"Medium Buy", 10000000, 0, 0.005, true},     // 0.01 SOL buy with 0.5% slippage
		{"Large Buy", 100000000, 0, 0.02, true},      // 0.1 SOL buy with 2% slippage
		{"Small Sell", 500000, 1, 0.01, true},        // 0.0005 SOL sell with 1% slippage
		{"Medium Sell", 5000000, 1, 0.005, true},     // 0.005 SOL sell with 0.5% slippage
		{"Large Sell", 50000000, 1, 0.02, true},      // 0.05 SOL sell with 2% slippage
		{"Quote without fees", 1000000, 0, 0.01, false}, // Quote excluding fees
	}

	fmt.Println("\n📋 Pool Information:")
	fmt.Printf("   Pool PDA: %s\n", poolPDA)
	fmt.Printf("   Base Mint: %s (SOL)\n", baseMint)
	fmt.Printf("   Quote Mint: %s (USDC)\n", quoteMint)

	// Check if pool exists before getting quotes
	fmt.Println("\n🔍 Checking Pool Status...")
	poolInfo, err := client.RPCClient.GetAccountInfo(ctx, poolPDA)
	if err != nil || poolInfo.Value == nil {
		fmt.Println("   ❌ Pool does not exist!")
		fmt.Println("   💡 Run 02_create_pool.go first to create the pool")
		fmt.Println("\n🔄 Showing example quote calculations (simulated)...")
		
		// Show simulated quotes for demonstration
		for i, scenario := range swapScenarios {
			fmt.Printf("\n%d. %s:\n", i+1, scenario.name)
			fmt.Printf("   Input: %d tokens\n", scenario.amount)
			fmt.Printf("   Type: %s\n", map[uint8]string{0: "Buy (Quote->Base)", 1: "Sell (Base->Quote)"}[scenario.swapType])
			fmt.Printf("   Slippage: %.1f%%\n", scenario.slippage*100)
			
			// Simulated calculation for demonstration
			simulatedOutput := uint64(float64(scenario.amount) * 0.95) // Assume 5% price impact
			simulatedMinOut := uint64(float64(simulatedOutput) * (1.0 - scenario.slippage))
			simulatedFee := uint64(float64(scenario.amount) * 0.0025) // 0.25% fee
			
			fmt.Printf("   → Simulated Output: %d tokens\n", simulatedOutput)
			fmt.Printf("   → Min Output (with slippage): %d tokens\n", simulatedMinOut)
			fmt.Printf("   → Estimated Fee: %d tokens\n", simulatedFee)
			fmt.Printf("   → Price Impact: ~5.00%%\n")
		}
		
		fmt.Println("\n✅ SwapQuote example completed (simulated)!")
		return
	}

	fmt.Println("   ✅ Pool exists! Getting real quotes...")

	// Get quotes for each scenario
	for i, scenario := range swapScenarios {
		fmt.Printf("\n%d. %s:\n", i+1, scenario.name)
		
		quoteParams := dbc.SwapQuoteParams{
			Pool:               poolPDA,
			BaseMint:           baseMint,
			QuoteMint:          quoteMint,
			Amount:             scenario.amount,
			SwapType:           scenario.swapType,
			Slippage:           scenario.slippage,
			IncludeFeesInQuote: scenario.includeFeesInQuote,
		}

		fmt.Printf("   Input: %d tokens\n", quoteParams.Amount)
		fmt.Printf("   Type: %s\n", map[uint8]string{0: "Buy (Quote->Base)", 1: "Sell (Base->Quote)"}[quoteParams.SwapType])
		fmt.Printf("   Slippage: %.1f%%\n", quoteParams.Slippage*100)
		fmt.Printf("   Include Fees: %v\n", quoteParams.IncludeFeesInQuote)

		// Get the quote
		result, err := client.SwapQuote(ctx, quoteParams)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		// Display quote results according to assessment requirements
		fmt.Printf("   → Swap Out Amount: %d tokens\n", result.SwapOutAmount)
		fmt.Printf("   → Min Swap Out Amount: %d tokens\n", result.MinSwapOutAmount)
		fmt.Printf("   → Price Impact: %.2f%%\n", result.PriceImpact)
		fmt.Printf("   → Fee: %d tokens\n", result.Fee)
		
		// Calculate additional metrics
		if scenario.swapType == 0 { // Buy
			effectivePrice := float64(quoteParams.Amount) / float64(result.SwapOutAmount)
			fmt.Printf("   → Effective Price: %.6f USDC per SOL\n", effectivePrice)
		} else { // Sell
			effectivePrice := float64(result.SwapOutAmount) / float64(quoteParams.Amount)
			fmt.Printf("   → Effective Price: %.6f USDC per SOL\n", effectivePrice)
		}
		
		slippageProtection := float64(result.SwapOutAmount - result.MinSwapOutAmount) / float64(result.SwapOutAmount) * 100
		fmt.Printf("   → Slippage Protection: %.2f%%\n", slippageProtection)
	}

	// Show example usage patterns
	fmt.Println("\n🔄 Example Usage Patterns:")
	fmt.Println("\n1. Pre-transaction Quote:")
	fmt.Println("   // Get quote before executing swap")
	fmt.Println("   quote, err := client.SwapQuote(ctx, quoteParams)")
	fmt.Println("   if err != nil { return err }")
	fmt.Println("   ")
	fmt.Println("   // Use quote results in swap parameters")
	fmt.Println("   swapParams.MinAmountOut = quote.MinSwapOutAmount")

	fmt.Println("\n2. Price Impact Check:")
	fmt.Println("   if quote.PriceImpact > 5.0 { // 5% threshold")
	fmt.Println("       fmt.Println(\"High price impact warning!\")")
	fmt.Println("   }")

	fmt.Println("\n3. Fee Comparison:")
	fmt.Println("   feePercentage := float64(quote.Fee) / float64(quote.SwapOutAmount + quote.Fee) * 100")
	fmt.Println("   fmt.Printf(\"Trading fee: %.2f%%\", feePercentage)")

	// Show batch quote example
	fmt.Println("\n📊 Batch Quote Analysis:")
	amounts := []uint64{1000000, 5000000, 10000000, 50000000, 100000000}
	
	fmt.Printf("%-12s %-12s %-12s %-12s %-12s\n", "Amount", "Output", "Min Output", "Price Impact", "Fee")
	fmt.Println("----------------------------------------------------------------")
	
	for _, amount := range amounts {
		quoteParams := dbc.SwapQuoteParams{
			Pool:               poolPDA,
			BaseMint:           baseMint,
			QuoteMint:          quoteMint,
			Amount:             amount,
			SwapType:           0, // Buy
			Slippage:           0.01, // 1%
			IncludeFeesInQuote: true,
		}

		result, err := client.SwapQuote(ctx, quoteParams)
		if err != nil {
			fmt.Printf("%-12d %-12s %-12s %-12s %-12s\n", amount, "Error", "Error", "Error", "Error")
			continue
		}

		fmt.Printf("%-12d %-12d %-12d %-11.2f%% %-12d\n", 
			amount, result.SwapOutAmount, result.MinSwapOutAmount, 
			result.PriceImpact, result.Fee)
	}

	fmt.Println("\n✅ SwapQuote example completed successfully!")
	fmt.Printf("🔗 Pool Explorer: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)
}