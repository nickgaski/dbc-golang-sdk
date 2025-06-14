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

	fmt.Println("📊 Meteora DBC - Pool Information Example")
	fmt.Println("=========================================")

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

	// Example: Get information for a SOL/USDC pool
	baseMint := dbc.GetNativeMint() // SOL
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC

	fmt.Printf("🔍 Fetching Pool Information:\n")
	fmt.Printf("   Base Mint: %s (SOL)\n", baseMint)
	fmt.Printf("   Quote Mint: %s (USDC)\n", quoteMint)

	// Derive pool accounts
	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)
	quoteVaultPDA := dbc.DeriveQuoteVaultPDA(poolPDA)
	configPDA := dbc.DeriveConfigPDA()

	fmt.Printf("\n📍 Derived Addresses:\n")
	fmt.Printf("   Pool: %s\n", poolPDA)
	fmt.Printf("   Base Vault: %s\n", baseVaultPDA)
	fmt.Printf("   Quote Vault: %s\n", quoteVaultPDA)
	fmt.Printf("   Config: %s\n", configPDA)

	// Check pool existence
	fmt.Println("\n🔎 Pool Status Check:")
	poolInfo, err := client.RPCClient.GetAccountInfo(ctx, poolPDA)
	if err != nil {
		fmt.Printf("   ❌ Error fetching pool: %v\n", err)
		return
	}

	if poolInfo.Value == nil {
		fmt.Println("   ⚠️  Pool does not exist")
		fmt.Println("   💡 Use the create_pool_and_swap.go example to create a pool")
		
		// Show what the pool structure would look like
		fmt.Println("\n📋 Pool Structure (when created):")
		fmt.Println("   • Base Reserve: Token balance in base vault")
		fmt.Println("   • Quote Reserve: Token balance in quote vault")
		fmt.Println("   • Current Price: sqrt_price in Q64.64 format")
		fmt.Println("   • Creator: Pool creator address")
		fmt.Println("   • Fees: Accumulated protocol and trading fees")
		fmt.Println("   • Migration Status: Whether pool has migrated")
		fmt.Println("   • Metrics: Trading volume and fee statistics")
		
	} else {
		fmt.Println("   ✅ Pool exists!")
		fmt.Printf("   📦 Data size: %d bytes\n", len(poolInfo.Value.Data.GetBinary()))
		fmt.Printf("   👤 Owner: %s\n", poolInfo.Value.Owner)
		
		// In a real implementation, you would deserialize the pool data here
		fmt.Println("\n📊 Pool Data (would be deserialized):")
		fmt.Println("   • Base Reserve: [needs deserialization]")
		fmt.Println("   • Quote Reserve: [needs deserialization]")
		fmt.Println("   • Current Price: [needs deserialization]")
		fmt.Println("   • Creator: [needs deserialization]")
		
		// Check vault balances
		fmt.Println("\n💰 Vault Balances:")
		baseBalance, err := client.RPCClient.GetTokenAccountBalance(ctx, baseVaultPDA, client.Commitment)
		if err != nil {
			fmt.Printf("   Base Vault: Error - %v\n", err)
		} else {
			fmt.Printf("   Base Vault: %s %s\n", baseBalance.Value.Amount, baseBalance.Value.UiAmountString)
		}
		
		quoteBalance, err := client.RPCClient.GetTokenAccountBalance(ctx, quoteVaultPDA, client.Commitment)
		if err != nil {
			fmt.Printf("   Quote Vault: Error - %v\n", err)
		} else {
			fmt.Printf("   Quote Vault: %s %s\n", quoteBalance.Value.Amount, quoteBalance.Value.UiAmountString)
		}
	}

	// Check configuration
	fmt.Println("\n⚙️  Configuration Status:")
	configInfo, err := client.RPCClient.GetAccountInfo(ctx, configPDA)
	if err != nil {
		fmt.Printf("   ❌ Error fetching config: %v\n", err)
	} else if configInfo.Value == nil {
		fmt.Println("   ⚠️  Configuration does not exist")
		fmt.Println("   💡 Configuration needs to be created by admin")
	} else {
		fmt.Println("   ✅ Configuration exists")
		fmt.Printf("   📦 Data size: %d bytes\n", len(configInfo.Value.Data.GetBinary()))
	}

	// Show mathematical functions
	fmt.Println("\n🧮 Mathematical Functions Available:")
	fmt.Println("   • Price calculation from reserves")
	fmt.Println("   • Swap quote calculation")
	fmt.Println("   • Fee calculation")
	fmt.Println("   • Price impact calculation")
	fmt.Println("   • Bonding curve mathematics")

	// Example price calculations (would work with real pool data)
	fmt.Println("\n📈 Price Calculation Example:")
	if poolInfo.Value != nil {
		fmt.Println("   💡 With real pool data, you could calculate:")
		fmt.Println("     • Current Price = sqrt_price²")
		fmt.Println("     • Price Impact = f(trade_size, liquidity)")
		fmt.Println("     • Expected Output = f(input, reserves, fees)")
	}

	fmt.Println("\n🔗 Explorer Links:")
	fmt.Printf("   • Pool: https://explorer.solana.com/address/%s?cluster=devnet\n", poolPDA)
	fmt.Printf("   • Base Vault: https://explorer.solana.com/address/%s?cluster=devnet\n", baseVaultPDA)
	fmt.Printf("   • Quote Vault: https://explorer.solana.com/address/%s?cluster=devnet\n", quoteVaultPDA)
	fmt.Printf("   • Config: https://explorer.solana.com/address/%s?cluster=devnet\n", configPDA)

	fmt.Println("\n✅ Pool information fetch completed!")
}