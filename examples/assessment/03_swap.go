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

// Example script for Swap as per assessment requirements
func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("💱 DBC Assessment - Swap Example")
	fmt.Println("=================================")

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

	// Define swap parameters according to assessment requirements
	baseMint := dbc.GetNativeMint() // SOL
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
	configPDA := dbc.DeriveConfigPDA()
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)
	quoteVaultPDA := dbc.DeriveQuoteVaultPDA(poolPDA)

	// Derive user token accounts
	userBaseAta, _, _ := solana.FindAssociatedTokenAddress(
		client.Payer.PublicKey(),
		baseMint,
	)
	userQuoteAta, _, _ := solana.FindAssociatedTokenAddress(
		client.Payer.PublicKey(),
		quoteMint,
	)

	swapAmount := uint64(1000000)      // 0.001 SOL (or 1 USDC with 6 decimals)
	minAmountOut := uint64(900000)     // Minimum amount out (accounting for slippage)
	
	swapParams := dbc.SwapParams{
		Config:           configPDA,
		Pool:             poolPDA,
		User:             client.Payer.PublicKey(),
		UserBaseAta:      userBaseAta,
		UserQuoteAta:     userQuoteAta,
		PoolBaseVault:    baseVaultPDA,
		PoolQuoteVault:   quoteVaultPDA,
		BaseMint:         baseMint,
		QuoteMint:        quoteMint,
		ReferralAccount:  client.Payer.PublicKey(), // Using self as referral
		Amount:           swapAmount,
		MinAmountOut:     minAmountOut,
		SwapType:         0, // 0 = Buy (Quote -> Base), 1 = Sell (Base -> Quote)
	}

	fmt.Println("\n📋 Swap Parameters:")
	fmt.Printf("   Pool: %s\n", swapParams.Pool)
	fmt.Printf("   User: %s\n", swapParams.User)
	fmt.Printf("   Base Mint: %s (SOL)\n", swapParams.BaseMint)
	fmt.Printf("   Quote Mint: %s (USDC)\n", swapParams.QuoteMint)
	fmt.Printf("   Swap Amount: %d\n", swapParams.Amount)
	fmt.Printf("   Min Amount Out: %d\n", swapParams.MinAmountOut)
	fmt.Printf("   Swap Type: %d (%s)\n", swapParams.SwapType, 
		map[uint8]string{0: "Buy (Quote->Base)", 1: "Sell (Base->Quote)"}[swapParams.SwapType])

	fmt.Println("\n🔗 Derived Accounts:")
	fmt.Printf("   User Base ATA: %s\n", userBaseAta)
	fmt.Printf("   User Quote ATA: %s\n", userQuoteAta)
	fmt.Printf("   Pool Base Vault: %s\n", baseVaultPDA)
	fmt.Printf("   Pool Quote Vault: %s\n", quoteVaultPDA)

	// Create the swap instruction
	fmt.Println("\n⚙️ Creating Swap Instruction...")
	
	instruction := client.Swap(swapParams)
	
	// Get instruction details
	accounts := instruction.Accounts()
	data, err := instruction.Data()
	if err != nil {
		log.Fatalf("Failed to get instruction data: %v", err)
	}

	fmt.Printf("✅ Swap instruction created successfully!\n")
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
		fmt.Println("   ❌ Pool does not exist!")
		fmt.Println("   💡 Run 02_create_pool.go first to create the pool")
	} else {
		fmt.Println("   ✅ Pool exists!")
	}

	// Check if config exists
	configInfo, err := client.RPCClient.GetAccountInfo(ctx, configPDA)
	if err != nil || configInfo.Value == nil {
		fmt.Println("   ❌ Configuration does not exist!")
		fmt.Println("   💡 Run 01_create_config.go first to create configuration")
	} else {
		fmt.Println("   ✅ Configuration exists!")
	}

	// Check user token accounts
	baseAtaInfo, err := client.RPCClient.GetAccountInfo(ctx, userBaseAta)
	if err != nil || baseAtaInfo.Value == nil {
		fmt.Println("   ⚠️  User Base ATA does not exist (will be created in transaction)")
	} else {
		fmt.Println("   ✅ User Base ATA exists!")
	}

	quoteAtaInfo, err := client.RPCClient.GetAccountInfo(ctx, userQuoteAta)
	if err != nil || quoteAtaInfo.Value == nil {
		fmt.Println("   ⚠️  User Quote ATA does not exist (will be created in transaction)")
	} else {
		fmt.Println("   ✅ User Quote ATA exists!")
	}

	// Show swap calculation details
	fmt.Println("\n🧮 Swap Calculation:")
	fmt.Printf("   Input Amount: %d\n", swapAmount)
	fmt.Printf("   Expected Min Output: %d\n", minAmountOut)
	fmt.Printf("   Slippage Tolerance: %.2f%%\n", 
		(1.0 - float64(minAmountOut)/float64(swapAmount)) * 100)

	// Show transaction execution example
	fmt.Println("\n🔄 Example Transaction Execution:")
	fmt.Println("   // Execute the swap")
	fmt.Println("   sig, err := client.ExecuteTransaction(ctx, instruction)")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatalf(\"Failed to execute transaction: %v\", err)")
	fmt.Println("   }")
	fmt.Println("   fmt.Printf(\"Transaction signature: %s\\n\", sig)")

	// Show alternative swap direction
	fmt.Println("\n🔄 Alternative Swap (Sell):")
	sellParams := swapParams
	sellParams.SwapType = 1 // Sell (Base -> Quote)
	sellParams.Amount = uint64(500000) // 0.0005 SOL
	sellParams.MinAmountOut = uint64(450000) // Minimum USDC out

	sellInstruction := client.Swap(sellParams)
	fmt.Printf("   Sell %d base tokens for minimum %d quote tokens\n", 
		sellParams.Amount, sellParams.MinAmountOut)
	fmt.Printf("   Instruction accounts: %d\n", len(sellInstruction.Accounts()))

	fmt.Println("\n✅ Swap example completed successfully!")
	fmt.Printf("🔗 Pool Explorer: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)

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