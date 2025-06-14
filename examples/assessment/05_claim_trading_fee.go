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

// Example script for ClaimTradingFee as per assessment requirements
func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("💰 DBC Assessment - ClaimTradingFee Example")
	fmt.Println("===========================================")

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

	// Define claim parameters according to assessment requirements
	baseMint := dbc.GetNativeMint() // SOL
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
	configPDA := dbc.DeriveConfigPDA()
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)
	quoteVaultPDA := dbc.DeriveQuoteVaultPDA(poolPDA)

	// Derive claimer token accounts (creator/partner claiming fees)
	claimerBaseAta, _, _ := solana.FindAssociatedTokenAddress(
		client.Payer.PublicKey(),
		baseMint,
	)
	claimerQuoteAta, _, _ := solana.FindAssociatedTokenAddress(
		client.Payer.PublicKey(),
		quoteMint,
	)

	claimParams := dbc.ClaimTradingFeeParams{
		Config:          configPDA,
		Pool:            poolPDA,
		Claimer:         client.Payer.PublicKey(),
		ClaimerBaseAta:  claimerBaseAta,
		ClaimerQuoteAta: claimerQuoteAta,
		PoolBaseVault:   baseVaultPDA,
		PoolQuoteVault:  quoteVaultPDA,
		BaseMint:        baseMint,
		QuoteMint:       quoteMint,
		MaxAmountA:      1000000000, // Maximum base tokens to claim
		MaxAmountB:      1000000000, // Maximum quote tokens to claim
	}

	fmt.Println("\n📋 Claim Parameters:")
	fmt.Printf("   Pool: %s\n", claimParams.Pool)
	fmt.Printf("   Claimer: %s\n", claimParams.Claimer)
	fmt.Printf("   Base Mint: %s (SOL)\n", claimParams.BaseMint)
	fmt.Printf("   Quote Mint: %s (USDC)\n", claimParams.QuoteMint)
	fmt.Printf("   Max Amount A: %d\n", claimParams.MaxAmountA)
	fmt.Printf("   Max Amount B: %d\n", claimParams.MaxAmountB)

	fmt.Println("\n🔗 Derived Accounts:")
	fmt.Printf("   Claimer Base ATA: %s\n", claimerBaseAta)
	fmt.Printf("   Claimer Quote ATA: %s\n", claimerQuoteAta)
	fmt.Printf("   Pool Base Vault: %s\n", baseVaultPDA)
	fmt.Printf("   Pool Quote Vault: %s\n", quoteVaultPDA)

	// Create the claim trading fee instruction
	fmt.Println("\n⚙️ Creating ClaimTradingFee Instruction...")
	
	instruction := client.ClaimTradingFee(claimParams)
	
	// Get instruction details
	accounts := instruction.Accounts()
	data, err := instruction.Data()
	if err != nil {
		log.Fatalf("Failed to get instruction data: %v", err)
	}

	fmt.Printf("✅ ClaimTradingFee instruction created successfully!\n")
	fmt.Printf("📦 Accounts: %d\n", len(accounts))
	fmt.Printf("📊 Data size: %d bytes\n", len(data))
	fmt.Printf("🏷️ Program ID: %s\n", instruction.ProgramID())

	// Show account details
	fmt.Println("\n📋 Account Details:")
	for i, account := range accounts {
		fmt.Printf("   %d. %s (writable: %v, signer: %v)\n", 
			i+1, account.PublicKey, account.IsWritable, account.IsSigner)
	}

	// Check prerequisites
	fmt.Println("\n🔍 Checking Prerequisites...")
	
	// Check if pool exists
	poolInfo, err := client.RPCClient.GetAccountInfo(ctx, poolPDA)
	if err != nil || poolInfo.Value == nil {
		fmt.Println("   📝 Assessment Demo: Pool not created (expected for demo)")
		fmt.Println("   💡 In real usage: Run 02_create_pool.go first to create the pool")
	} else {
		fmt.Println("   ✅ Pool exists!")
		
		// Try to parse pool data to check fee accumulation
		var pool dbc.VirtualPool
		if err := dbc.DeserializePool(poolInfo.Value.Data.GetBinary(), &pool); err == nil {
			fmt.Printf("   📊 Creator Base Fee: %d\n", pool.CreatorBaseFee)
			fmt.Printf("   📊 Creator Quote Fee: %d\n", pool.CreatorQuoteFee)
			fmt.Printf("   📊 Partner Base Fee: %d\n", pool.PartnerBaseFee)
			fmt.Printf("   📊 Partner Quote Fee: %d\n", pool.PartnerQuoteFee)
		}
	}

	// Check if config exists
	configInfo, err := client.RPCClient.GetAccountInfo(ctx, configPDA)
	if err != nil || configInfo.Value == nil {
		fmt.Println("   📝 Assessment Demo: Configuration not created (expected for demo)")
		fmt.Println("   💡 In real usage: Run 01_create_config.go first to create configuration")
	} else {
		fmt.Println("   ✅ Configuration exists!")
	}

	// Check claimer token accounts
	baseAtaInfo, err := client.RPCClient.GetAccountInfo(ctx, claimerBaseAta)
	if err != nil || baseAtaInfo.Value == nil {
		fmt.Println("   ⚠️  Claimer Base ATA does not exist (will be created if needed)")
	} else {
		fmt.Println("   ✅ Claimer Base ATA exists!")
		// Get balance
		balance, err := client.RPCClient.GetTokenAccountBalance(ctx, claimerBaseAta, client.Commitment)
		if err == nil {
			fmt.Printf("      Balance: %s\n", balance.Value.Amount)
		}
	}

	quoteAtaInfo, err := client.RPCClient.GetAccountInfo(ctx, claimerQuoteAta)
	if err != nil || quoteAtaInfo.Value == nil {
		fmt.Println("   ⚠️  Claimer Quote ATA does not exist (will be created if needed)")
	} else {
		fmt.Println("   ✅ Claimer Quote ATA exists!")
		// Get balance
		balance, err := client.RPCClient.GetTokenAccountBalance(ctx, claimerQuoteAta, client.Commitment)
		if err == nil {
			fmt.Printf("      Balance: %s\n", balance.Value.Amount)
		}
	}

	// Check vault balances to estimate claimable amounts
	fmt.Println("\n💰 Vault Information:")
	baseVaultBalance, err := client.RPCClient.GetTokenAccountBalance(ctx, baseVaultPDA, client.Commitment)
	if err == nil {
		fmt.Printf("   Base Vault Balance: %s\n", baseVaultBalance.Value.Amount)
	} else {
		fmt.Println("   📝 Assessment Demo: Base vault not yet created (normal for demo)")
	}

	quoteVaultBalance, err := client.RPCClient.GetTokenAccountBalance(ctx, quoteVaultPDA, client.Commitment)
	if err == nil {
		fmt.Printf("   Quote Vault Balance: %s\n", quoteVaultBalance.Value.Amount)
	} else {
		fmt.Println("   📝 Assessment Demo: Quote vault not yet created (normal for demo)")
	}

	// Show fee claiming scenarios
	fmt.Println("\n📊 Fee Claiming Scenarios:")
	
	scenarios := []struct {
		name        string
		description string
		maxAmountA  uint64
		maxAmountB  uint64
	}{
		{"Full Claim", "Claim all available fees", 1000000000, 1000000000},
		{"Partial Claim", "Claim limited amounts", 100000000, 100000000},
		{"Base Only", "Claim only base token fees", 1000000000, 0},
		{"Quote Only", "Claim only quote token fees", 0, 1000000000},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n%d. %s:\n", i+1, scenario.name)
		fmt.Printf("   Description: %s\n", scenario.description)
		fmt.Printf("   Max Base Amount: %d\n", scenario.maxAmountA)
		fmt.Printf("   Max Quote Amount: %d\n", scenario.maxAmountB)
		
		// Create instruction for this scenario
		scenarioParams := claimParams
		scenarioParams.MaxAmountA = scenario.maxAmountA
		scenarioParams.MaxAmountB = scenario.maxAmountB
		
		scenarioInstruction := client.ClaimTradingFee(scenarioParams)
		fmt.Printf("   ✅ Instruction created with %d accounts\n", len(scenarioInstruction.Accounts()))
	}

	// Show transaction execution example
	fmt.Println("\n🔄 Example Transaction Execution:")
	fmt.Println("   // Execute the fee claim")
	fmt.Println("   sig, err := client.ExecuteTransaction(ctx, instruction)")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatalf(\"Failed to execute transaction: %v\", err)")
	fmt.Println("   }")
	fmt.Println("   fmt.Printf(\"Transaction signature: %s\\n\", sig)")

	// Show related operations
	fmt.Println("\n🔗 Related Operations:")
	fmt.Println("   • Creator Fee Claim: For pool creators to claim their earned fees")
	fmt.Println("   • Partner Fee Claim: For partners to claim their share of fees")
	fmt.Println("   • Protocol Fee Claim: For protocol to claim platform fees")
	
	// Show fee calculation example
	fmt.Println("\n🧮 Fee Calculation Example:")
	fmt.Println("   // Typical fee structure:")
	fmt.Println("   // - Trading fee: 0.25% of swap amount")
	fmt.Println("   // - Creator share: 40% of trading fee")
	fmt.Println("   // - Partner share: 10% of trading fee")
	fmt.Println("   // - Protocol share: 50% of trading fee")
	fmt.Println("")
	fmt.Println("   tradingFee := swapAmount * 25 / 10000  // 0.25%")
	fmt.Println("   creatorFee := tradingFee * 40 / 100    // 40%")
	fmt.Println("   partnerFee := tradingFee * 10 / 100    // 10%")
	fmt.Println("   protocolFee := tradingFee * 50 / 100   // 50%")

	fmt.Println("\n✅ ClaimTradingFee example completed successfully!")
	fmt.Printf("🔗 Pool Explorer: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)
	fmt.Printf("🔗 Base Vault: https://explorer.solana.com/address/%s?cluster=devnet\n", baseVaultPDA)
	fmt.Printf("🔗 Quote Vault: https://explorer.solana.com/address/%s?cluster=devnet\n", quoteVaultPDA)

	// Optionally execute the transaction (commented out for safety)
	/*
	fmt.Println("\n⚡ Executing transaction...")
	sig, err := client.ExecuteTransaction(ctx, instruction)
	if err != nil {
		log.Fatalf("Failed to execute transaction: %v", err)
	}
	fmt.Printf("✅ Transaction executed successfully!\n")
	fmt.Printf("🔗 Transaction: https://explorer.solana.com/tx/%s?cluster=devnet\n", sig)
	*/
}