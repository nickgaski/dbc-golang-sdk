package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dbc-golang-sdk/pkg/dbc"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("🔍 DBC Program Deployment Verification")
	fmt.Println("=====================================")

	// Create RPC client
	client := rpc.New("https://api.devnet.solana.com")
	ctx := context.Background()

	// Verify DBC program deployment
	programID := dbc.GetDynamicBondingCurveProgramID()
	fmt.Printf("📋 Checking DBC Program: %s\n", programID)

	programInfo, err := client.GetAccountInfo(ctx, programID)
	if err != nil {
		log.Fatalf("Failed to get program info: %v", err)
	}

	if programInfo.Value == nil {
		log.Fatalf("Program not found on devnet")
	}

	fmt.Printf("✅ Program deployed and executable: %v\n", programInfo.Value.Executable)
	fmt.Printf("📊 Program owner: %s\n", programInfo.Value.Owner)
	fmt.Printf("💰 Program balance: %.6f SOL\n", float64(programInfo.Value.Lamports)/1e9)
	fmt.Printf("📏 Program size: %d bytes\n", len(programInfo.Value.Data.GetBinary()))

	// Check if program is upgradeable
	if programInfo.Value.Owner.Equals(solana.MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111")) {
		fmt.Println("🔄 Program is upgradeable")
	} else {
		fmt.Println("🔒 Program is immutable")
	}

	// Get keypair for testing
	keypairFile := os.Getenv("KEYPAIR_FILE")
	if keypairFile == "" {
		keypairFile = os.Getenv("HOME") + "/.config/solana/devnet-keypair.json"
	}

	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(keypairFile)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	fmt.Printf("🔑 Using account: %s\n", privateKey.PublicKey())

	// Test instruction creation (without execution)
	fmt.Println("\n🧪 Testing Instruction Creation")
	fmt.Println("===============================")

	// Test CreateConfig instruction
	configPDA := dbc.DeriveConfigPDA()
	fmt.Printf("📍 Config PDA: %s\n", configPDA)

	configParams := dbc.ConfigParams{
		Admin: privateKey.PublicKey(),
		BaseFee: dbc.BaseFeeConfig{
			CliffFeeNumerator: 1000,
			PeriodFrequency:   3600,
			ReductionFactor:   9000,
			NumberOfPeriod:    100,
			FeeSchedulerMode:  0,
		},
		DynamicFee: dbc.DynamicFeeConfig{
			Initialized:              1,
			MaxVolatilityAccumulator: 2000,
			VariableFeeControl:       500,
			BinStep:                  25,
			FilterPeriod:             30,
			DecayPeriod:              600,
			ReductionFactor:          5000,
		},
		ProtocolFeePercent:     25,
		ReferralFeePercent:     10,
		DefaultReferralAccount: privateKey.PublicKey(),
		MigrationOption:        uint8(dbc.MigrationOptionMeteoraDamm),
		PlatformFeeRecipient:   privateKey.PublicKey(),
	}

	config := dbc.DevnetConfig()
	dbcClient := dbc.NewDBCClientWithConfig(config, privateKey)

	// Test instruction creation
	instruction := dbcClient.CreateConfig(configParams)
	accounts := instruction.Accounts()
	data, err := instruction.Data()
	if err != nil {
		log.Printf("⚠️ Error getting instruction data: %v", err)
	} else {
		fmt.Printf("✅ CreateConfig instruction created successfully\n")
		fmt.Printf("   📦 Accounts: %d\n", len(accounts))
		fmt.Printf("   📊 Data size: %d bytes\n", len(data))
		fmt.Printf("   🏷️ Program ID: %s\n", instruction.ProgramID())
	}

	// Test other PDAs
	fmt.Println("\n🗺️ Testing PDA Derivation")
	fmt.Println("=========================")

	baseMint := dbc.GetNativeMint()
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	poolPDA := dbc.DerivePoolPDA(baseMint, quoteMint)
	baseVaultPDA := dbc.DeriveBaseVaultPDA(poolPDA)
	quoteVaultPDA := dbc.DeriveQuoteVaultPDA(poolPDA)
	poolAuthorityPDA := dbc.DerivePoolAuthorityPDA()
	eventAuthorityPDA := dbc.DeriveEventAuthorityPDA()

	fmt.Printf("📍 Pool PDA: %s\n", poolPDA)
	fmt.Printf("📍 Base Vault PDA: %s\n", baseVaultPDA)
	fmt.Printf("📍 Quote Vault PDA: %s\n", quoteVaultPDA)
	fmt.Printf("📍 Pool Authority PDA: %s\n", poolAuthorityPDA)
	fmt.Printf("📍 Event Authority PDA: %s\n", eventAuthorityPDA)

	// Check if any accounts exist
	fmt.Println("\n🔍 Checking Account Existence")
	fmt.Println("=============================")

	accounts_to_check := map[string]solana.PublicKey{
		"Config":         configPDA,
		"Pool":           poolPDA,
		"Base Vault":     baseVaultPDA,
		"Quote Vault":    quoteVaultPDA,
		"Pool Authority": poolAuthorityPDA,
		"Event Authority": eventAuthorityPDA,
	}

	for name, pubkey := range accounts_to_check {
		info, err := client.GetAccountInfo(ctx, pubkey)
		if err == nil && info.Value != nil {
			fmt.Printf("✅ %s exists: %d bytes\n", name, len(info.Value.Data.GetBinary()))
		} else {
			fmt.Printf("❌ %s does not exist\n", name)
		}
	}

	// Test instruction discriminators
	fmt.Println("\n🏷️ Instruction Discriminators")
	fmt.Println("=============================")

	discriminators := map[string][8]byte{
		"CreateConfig":                        dbc.CreateConfigDiscriminator,
		"InitializeVirtualPoolWithSplToken":   dbc.InitializeVirtualPoolWithSplTokenDiscriminator,
		"Swap":                               dbc.SwapDiscriminator,
		"ClaimCreatorTradingFee":             dbc.ClaimCreatorTradingFeeDiscriminator,
		"WithdrawLeftover":                   dbc.WithdrawLeftoverDiscriminator,
		"ClaimTradingFee":                    dbc.ClaimTradingFeeDiscriminator,
		"MigrateMeteoraDamm":                 dbc.MigrateMeteoraDammDiscriminator,
		"MigrationDammV2":                    dbc.MigrationDammV2Discriminator,
	}

	for name, disc := range discriminators {
		fmt.Printf("🏷️ %s: %x\n", name, disc)
	}

	fmt.Println("\n✅ Program deployment verification completed!")
	fmt.Printf("🔗 Explorer: https://explorer.solana.com/address/%s?cluster=devnet\n", programID)
}