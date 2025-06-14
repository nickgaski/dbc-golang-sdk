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

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("🏊 Meteora DBC - Create Pool and Swap Example")
	fmt.Println("=============================================")

	// Get private key
	privateKeyStr := os.Getenv("PRIVATE_KEY")
	if privateKeyStr == "" {
		log.Fatal("PRIVATE_KEY environment variable not set")
	}

	privateKey, err := solana.PrivateKeyFromBase58(privateKeyStr)
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
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

	// Example token mints (using devnet tokens)
	baseMint := solana.MustPublicKeyFromBase58(dbc.NativeMint) // SOL
	// For this example, let's create a new SPL token as the base mint
	// In a real scenario, you'd have your token mint ready
	
	quoteMint := dbc.GetNativeMint() // WSOL
	
	fmt.Printf("📋 Pool Configuration:\n")
	fmt.Printf("   Base Mint: %s (SOL)\n", baseMint)
	fmt.Printf("   Quote Mint: %s (WSOL)\n", quoteMint)

	// Step 1: Create configuration if it doesn't exist
	fmt.Println("\n1️⃣ Creating DBC Configuration...")
	
	configPDA := dbc.DeriveConfigPDA()
	fmt.Printf("   Config PDA: %s\n", configPDA)

	// Check if config already exists
	configInfo, err := client.RPCClient.GetAccountInfo(ctx, configPDA)
	if err != nil || configInfo.Value == nil {
		fmt.Println("   ⚠️  Config not found, would need to create it")
		fmt.Println("   (Skipping config creation - requires admin privileges)")
	} else {
		fmt.Println("   ✅ Config already exists")
	}

	// Step 2: Derive pool accounts
	fmt.Println("\n2️⃣ Deriving Pool Accounts...")
	
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)
	quoteVaultPDA := dbc.DeriveQuoteVaultPDA(poolPDA)
	
	fmt.Printf("   Pool PDA: %s\n", poolPDA)
	fmt.Printf("   Base Vault: %s\n", baseVaultPDA)
	fmt.Printf("   Quote Vault: %s\n", quoteVaultPDA)

	// Step 3: Check if pool exists
	fmt.Println("\n3️⃣ Checking Pool Status...")
	
	poolInfo, err := client.RPCClient.GetAccountInfo(ctx, poolPDA)
	if err != nil || poolInfo.Value == nil {
		fmt.Println("   ⚠️  Pool does not exist")
		fmt.Println("   💡 Would need to create pool with InitializeVirtualPoolWithSplToken")
		
		// Show how the instruction would be created
		fmt.Println("\n   📝 Pool Creation Instruction Example:")
		
		// For metadata
		metadataPDA := dbc.DeriveMetadataPDA(baseMint)
		
		instruction := dbc.InitializeVirtualPoolWithSplTokenInstruction(
			configPDA,
			client.Payer.PublicKey(), // pool creator
			baseMint,
			quoteMint,
			poolPDA,
			baseVaultPDA,
			quoteVaultPDA,
			metadataPDA,
			client.Payer.PublicKey(), // payer
			"My Test Token",          // name
			"MTT",                    // symbol
			"https://example.com/metadata.json", // uri
		)
		
		accounts := instruction.Accounts()
		data, _ := instruction.Data()
		fmt.Printf("   ✅ Instruction created with %d accounts\n", len(accounts))
		fmt.Printf("   📦 Data size: %d bytes\n", len(data))
		
	} else {
		fmt.Println("   ✅ Pool exists!")
		
		// Step 4: Test swap quote
		fmt.Println("\n4️⃣ Testing Swap Quote...")
		
		swapAmount := uint64(1000000) // 0.001 SOL
		
		// This would fail since we can't actually fetch pool data without the program deployed
		fmt.Printf("   💱 Quote for %d lamports\n", swapAmount)
		fmt.Println("   ⚠️  Cannot fetch real quote without deployed program")
		
		// Show how swap instruction would be created
		fmt.Println("\n   📝 Swap Instruction Example:")
		
		// These would be the user's token accounts
		userBaseAccount := client.Payer.PublicKey() // placeholder
		userQuoteAccount := client.Payer.PublicKey() // placeholder
		referralAccount := client.Payer.PublicKey() // placeholder
		
		swapInstruction := dbc.SwapInstruction(
			configPDA,
			poolPDA,
			userBaseAccount,
			userQuoteAccount,
			baseVaultPDA,
			quoteVaultPDA,
			baseMint,
			quoteMint,
			client.Payer.PublicKey(),
			referralAccount,
			swapAmount,
			0, // min out (would calculate from quote)
		)
		
		swapAccounts := swapInstruction.Accounts()
		swapData, _ := swapInstruction.Data()
		fmt.Printf("   ✅ Swap instruction created with %d accounts\n", len(swapAccounts))
		fmt.Printf("   📦 Data size: %d bytes\n", len(swapData))
	}

	// Step 5: Show other available operations
	fmt.Println("\n5️⃣ Other Available Operations:")
	
	fmt.Println("   🎯 Creator Functions:")
	claimFeeIx := dbc.ClaimCreatorTradingFeeInstruction(
		configPDA, poolPDA, baseVaultPDA, quoteVaultPDA,
		baseMint, quoteMint,
		client.Payer.PublicKey(), client.Payer.PublicKey(), // creator token accounts
		client.Payer.PublicKey(), // creator
	)
	claimAccounts := claimFeeIx.Accounts()
	fmt.Printf("     • Claim Trading Fee: %d accounts\n", len(claimAccounts))
	
	withdrawIx := dbc.WithdrawLeftoverInstruction(
		configPDA, poolPDA, baseVaultPDA, baseMint,
		client.Payer.PublicKey(), client.Payer.PublicKey(), // leftover receiver
		client.Payer.PublicKey(), // creator
	)
	withdrawAccounts := withdrawIx.Accounts()
	fmt.Printf("     • Withdraw Leftover: %d accounts\n", len(withdrawAccounts))

	fmt.Println("\n✅ Example completed successfully!")
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("   1. Deploy the DBC program to devnet")
	fmt.Println("   2. Create a configuration account")
	fmt.Println("   3. Initialize a pool with real tokens")
	fmt.Println("   4. Execute swaps and other operations")
	fmt.Println("\n🔗 Useful Links:")
	fmt.Printf("   • Config PDA: https://explorer.solana.com/address/%s?cluster=devnet\n", configPDA)
	fmt.Printf("   • Pool PDA: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)
}