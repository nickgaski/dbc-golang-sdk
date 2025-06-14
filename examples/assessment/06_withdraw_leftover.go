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

// Example script for WithdrawLeftover as per assessment requirements
func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("🏦 DBC Assessment - WithdrawLeftover Example")
	fmt.Println("============================================")

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

	// Define withdraw parameters according to assessment requirements
	baseMint := dbc.GetNativeMint() // SOL
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
	configPDA := dbc.DeriveConfigPDA()
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)

	// Derive creator token account for receiving leftover tokens
	creatorBaseAta, _, _ := solana.FindAssociatedTokenAddress(
		client.Payer.PublicKey(),
		baseMint,
	)

	withdrawParams := dbc.WithdrawLeftoverParams{
		Config:           configPDA,
		Pool:             poolPDA,
		Creator:          client.Payer.PublicKey(),
		CreatorBaseAta:   creatorBaseAta,
		PoolBaseVault:    baseVaultPDA,
		BaseMint:         baseMint,
		LeftoverReceiver: client.Payer.PublicKey(), // Creator receives leftover tokens
	}

	fmt.Println("\n📋 Withdraw Parameters:")
	fmt.Printf("   Pool: %s\n", withdrawParams.Pool)
	fmt.Printf("   Creator: %s\n", withdrawParams.Creator)
	fmt.Printf("   Base Mint: %s (SOL)\n", withdrawParams.BaseMint)
	fmt.Printf("   Leftover Receiver: %s\n", withdrawParams.LeftoverReceiver)

	fmt.Println("\n🔗 Derived Accounts:")
	fmt.Printf("   Creator Base ATA: %s\n", creatorBaseAta)
	fmt.Printf("   Pool Base Vault: %s\n", baseVaultPDA)

	// Create the withdraw leftover instruction
	fmt.Println("\n⚙️ Creating WithdrawLeftover Instruction...")
	
	instruction := client.WithdrawLeftover(withdrawParams)
	
	// Get instruction details
	accounts := instruction.Accounts()
	data, err := instruction.Data()
	if err != nil {
		log.Fatalf("Failed to get instruction data: %v", err)
	}

	fmt.Printf("✅ WithdrawLeftover instruction created successfully!\n")
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
		
		// Try to parse pool data to check withdrawal status
		var pool dbc.VirtualPool
		if err := dbc.DeserializePool(poolInfo.Value.Data.GetBinary(), &pool); err == nil {
			fmt.Printf("   📊 Is Withdraw Leftover: %v\n", pool.IsWithdrawLeftover == 1)
			fmt.Printf("   📊 Migration Status: %v\n", pool.IsMigrated == 1)
			fmt.Printf("   📊 Pool Type: %d\n", pool.PoolType)
			
			// Check if creator matches
			if pool.Creator.Equals(client.Payer.PublicKey()) {
				fmt.Println("   ✅ You are the pool creator!")
			} else {
				fmt.Printf("   ⚠️  Pool creator is %s (you are %s)\n", pool.Creator, client.Payer.PublicKey())
			}
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

	// Check creator token account
	baseAtaInfo, err := client.RPCClient.GetAccountInfo(ctx, creatorBaseAta)
	if err != nil || baseAtaInfo.Value == nil {
		fmt.Println("   ⚠️  Creator Base ATA does not exist (will be created if needed)")
	} else {
		fmt.Println("   ✅ Creator Base ATA exists!")
		// Get balance
		balance, err := client.RPCClient.GetTokenAccountBalance(ctx, creatorBaseAta, client.Commitment)
		if err == nil {
			fmt.Printf("      Current Balance: %s\n", balance.Value.Amount)
		}
	}

	// Check vault balance to estimate leftover amount
	fmt.Println("\n💰 Vault Information:")
	baseVaultBalance, err := client.RPCClient.GetTokenAccountBalance(ctx, baseVaultPDA, client.Commitment)
	if err == nil {
		fmt.Printf("   Base Vault Balance: %s\n", baseVaultBalance.Value.Amount)
		fmt.Printf("   💡 This represents the total tokens in the vault\n")
		fmt.Printf("   💡 Leftover = Vault Balance - Active Liquidity\n")
	} else {
		fmt.Println("   📝 Assessment Demo: Base vault not yet created (normal for demo)")
	}

	// Show withdrawal conditions
	fmt.Println("\n🔒 Withdrawal Conditions:")
	fmt.Println("   ✓ Must be the pool creator")
	fmt.Println("   ✓ Pool must be in withdrawable state")
	fmt.Println("   ✓ Migration or bonding curve completion may be required")
	fmt.Println("   ✓ Leftover tokens must be available")

	// Show withdrawal scenarios
	fmt.Println("\n📊 Withdrawal Scenarios:")
	
	scenarios := []struct {
		name        string
		description string
		condition   string
	}{
		{
			"Post-Migration Leftover", 
			"Tokens remaining after successful migration to DAMM",
			"Pool.IsMigrated == true && leftover_balance > 0",
		},
		{
			"Bonding Curve Completion", 
			"Tokens remaining after bonding curve reaches end",
			"Pool.FinishCurveTimestamp > 0 && current_time > finish_time",
		},
		{
			"Emergency Withdrawal", 
			"Creator withdraws remaining tokens in special conditions",
			"Pool.IsWithdrawLeftover == true",
		},
		{
			"Partial Withdrawal", 
			"Withdraw specific amounts during certain phases",
			"Migration progress allows partial withdrawal",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n%d. %s:\n", i+1, scenario.name)
		fmt.Printf("   Description: %s\n", scenario.description)
		fmt.Printf("   Condition: %s\n", scenario.condition)
	}

	// Show transaction execution example
	fmt.Println("\n🔄 Example Transaction Execution:")
	fmt.Println("   // Execute the leftover withdrawal")
	fmt.Println("   sig, err := client.ExecuteTransaction(ctx, instruction)")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatalf(\"Failed to execute transaction: %v\", err)")
	fmt.Println("   }")
	fmt.Println("   fmt.Printf(\"Transaction signature: %s\\n\", sig)")

	// Show pre-withdrawal checks
	fmt.Println("\n🔍 Pre-Withdrawal Checks:")
	fmt.Println("   // Check if withdrawal is allowed")
	fmt.Println("   poolData, err := client.GetPoolData(ctx, poolPDA)")
	fmt.Println("   if err != nil { return err }")
	fmt.Println("   ")
	fmt.Println("   if !poolData.IsWithdrawLeftover {")
	fmt.Println("       return fmt.Errorf(\"withdrawal not allowed\")")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   if !poolData.Creator.Equals(myPublicKey) {")
	fmt.Println("       return fmt.Errorf(\"only creator can withdraw\")")
	fmt.Println("   }")

	// Show related operations
	fmt.Println("\n🔗 Related Operations:")
	fmt.Println("   • Migration to DAMM: Complete bonding curve and migrate to AMM")
	fmt.Println("   • LP Token Claims: Claim creator/partner LP tokens after migration")
	fmt.Println("   • Fee Claims: Claim accumulated trading fees")
	fmt.Println("   • Emergency Functions: Special withdrawal in emergency conditions")

	// Show calculation example
	fmt.Println("\n🧮 Leftover Calculation Example:")
	fmt.Println("   // Leftover amount calculation")
	fmt.Println("   totalVaultBalance := getVaultBalance(baseVault)")
	fmt.Println("   activeLiquidity := getActiveLiquidity(pool)")
	fmt.Println("   reservedForFees := getReservedFees(pool)")
	fmt.Println("   ")
	fmt.Println("   leftoverAmount := totalVaultBalance - activeLiquidity - reservedForFees")
	fmt.Println("   ")
	fmt.Println("   if leftoverAmount > 0 {")
	fmt.Println("       // Execute withdrawal")
	fmt.Println("   }")

	fmt.Println("\n✅ WithdrawLeftover example completed successfully!")
	fmt.Printf("🔗 Pool Explorer: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)
	fmt.Printf("🔗 Base Vault: https://explorer.solana.com/address/%s?cluster=devnet\n", baseVaultPDA)
	fmt.Printf("🔗 Creator ATA: https://explorer.solana.com/address/%s?cluster=devnet\n", creatorBaseAta)

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