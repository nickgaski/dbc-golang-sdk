package dbc

import (
	"testing"

	"github.com/gagliardetto/solana-go"
)

func TestCreateConfigInstruction(t *testing.T) {
	// Test data - use valid base58 public keys (32 bytes each)
	config := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	payer := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	feeClaimer := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")
	leftoverReceiver := solana.MustPublicKeyFromBase58("GJBKevjCRGwBgj42Wz2Z7sYLHQpB9DYQbLgfYP6tM9vK")
	
	poolConfig := PoolConfig{
		QuoteMint:        GetNativeMint(),
		FeeClaimer:       feeClaimer,
		LeftoverReceiver: leftoverReceiver,
		PoolFees: PoolFeesConfig{
			ProtocolFeePercent: 25,
			ReferralFeePercent: 10,
		},
		MigrationOption: 1,
	}

	instruction := CreateConfigInstruction(config, payer, feeClaimer, leftoverReceiver, poolConfig)

	// Verify instruction structure
	programID := instruction.ProgramID()
	if !programID.Equals(GetDynamicBondingCurveProgramID()) {
		t.Errorf("Expected program ID %s, got %s", GetDynamicBondingCurveProgramID(), programID)
	}

	// Verify accounts count
	expectedAccountCount := 3 // config, payer, system_program
	accounts := instruction.Accounts()
	if len(accounts) != expectedAccountCount {
		t.Errorf("Expected %d accounts, got %d", expectedAccountCount, len(accounts))
	}

	// Verify instruction data is not empty and starts with correct discriminator
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Failed to get instruction data: %v", err)
	}
	if len(data) < 8 {
		t.Errorf("Instruction data too short: %d bytes", len(data))
	}

	// Check that discriminator matches
	expectedDiscriminator := CreateConfigDiscriminator
	for i := 0; i < 8; i++ {
		if data[i] != expectedDiscriminator[i] {
			t.Errorf("Discriminator byte %d: expected %02x, got %02x", 
				i, expectedDiscriminator[i], data[i])
		}
	}

	t.Logf("CreateConfig instruction data length: %d bytes", len(data))
	t.Logf("Discriminator: %x", data[:8])
}

func TestSwapInstruction(t *testing.T) {
	// Test data - use valid base58 public keys
	config := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	pool := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	userInputAccount := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")
	userOutputAccount := solana.MustPublicKeyFromBase58("GJBKevjCRGwBgj42Wz2Z7sYLHQpB9DYQbLgfYP6tM9vK")
	baseVault := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	quoteVault := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	baseMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	quoteMint := GetNativeMint()
	payer := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	referralAccount := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")

	instruction := SwapInstruction(
		config, pool, userInputAccount, userOutputAccount,
		baseVault, quoteVault, baseMint, quoteMint, payer, referralAccount,
		1000000, 900000, // amountIn, minOut
	)

	// Verify instruction structure
	programID := instruction.ProgramID()
	if !programID.Equals(GetDynamicBondingCurveProgramID()) {
		t.Errorf("Expected program ID %s, got %s", GetDynamicBondingCurveProgramID(), programID)
	}

	// Verify instruction data is not empty and starts with correct discriminator
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Failed to get instruction data: %v", err)
	}
	if len(data) < 8 {
		t.Errorf("Instruction data too short: %d bytes", len(data))
	}

	// Check that discriminator matches
	expectedDiscriminator := SwapDiscriminator
	for i := 0; i < 8; i++ {
		if data[i] != expectedDiscriminator[i] {
			t.Errorf("Discriminator byte %d: expected %02x, got %02x", 
				i, expectedDiscriminator[i], data[i])
		}
	}

	t.Logf("Swap instruction data length: %d bytes", len(data))
	t.Logf("Discriminator: %x", data[:8])
}

func TestInitializeVirtualPoolWithSplTokenInstruction(t *testing.T) {
	// Test data - use valid base58 public keys
	config := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	poolCreator := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	baseMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	quoteMint := GetNativeMint()
	pool := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")
	baseVault := solana.MustPublicKeyFromBase58("GJBKevjCRGwBgj42Wz2Z7sYLHQpB9DYQbLgfYP6tM9vK")
	quoteVault := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	mintMetadata := solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	payer := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")

	instruction := InitializeVirtualPoolWithSplTokenInstruction(
		config, poolCreator, baseMint, quoteMint, pool, baseVault, quoteVault,
		mintMetadata, payer, "Test Token", "TEST", "https://example.com",
	)

	// Verify instruction structure
	programID := instruction.ProgramID()
	if !programID.Equals(GetDynamicBondingCurveProgramID()) {
		t.Errorf("Expected program ID %s, got %s", GetDynamicBondingCurveProgramID(), programID)
	}

	// Verify instruction data is not empty and starts with correct discriminator
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Failed to get instruction data: %v", err)
	}
	if len(data) < 8 {
		t.Errorf("Instruction data too short: %d bytes", len(data))
	}

	// Check that discriminator matches
	expectedDiscriminator := InitializeVirtualPoolWithSplTokenDiscriminator
	for i := 0; i < 8; i++ {
		if data[i] != expectedDiscriminator[i] {
			t.Errorf("Discriminator byte %d: expected %02x, got %02x", 
				i, expectedDiscriminator[i], data[i])
		}
	}

	t.Logf("InitializeVirtualPoolWithSplToken instruction data length: %d bytes", len(data))
	t.Logf("Discriminator: %x", data[:8])
}

func TestClaimCreatorTradingFeeInstruction(t *testing.T) {
	// Test data - use valid base58 public keys
	config := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	pool := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	baseVault := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")
	quoteVault := solana.MustPublicKeyFromBase58("GJBKevjCRGwBgj42Wz2Z7sYLHQpB9DYQbLgfYP6tM9vK")
	baseMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	quoteMint := GetNativeMint()
	creatorBaseTokenAccount := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	creatorQuoteTokenAccount := solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	creator := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")

	instruction := ClaimCreatorTradingFeeInstruction(
		config, pool, baseVault, quoteVault, baseMint, quoteMint,
		creatorBaseTokenAccount, creatorQuoteTokenAccount, creator,
	)

	// Verify instruction structure
	programID := instruction.ProgramID()
	if !programID.Equals(GetDynamicBondingCurveProgramID()) {
		t.Errorf("Expected program ID %s, got %s", GetDynamicBondingCurveProgramID(), programID)
	}

	// Verify instruction data is not empty and starts with correct discriminator
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Failed to get instruction data: %v", err)
	}
	if len(data) < 8 {
		t.Errorf("Instruction data too short: %d bytes", len(data))
	}

	// Check that discriminator matches
	expectedDiscriminator := ClaimCreatorTradingFeeDiscriminator
	for i := 0; i < 8; i++ {
		if data[i] != expectedDiscriminator[i] {
			t.Errorf("Discriminator byte %d: expected %02x, got %02x", 
				i, expectedDiscriminator[i], data[i])
		}
	}

	t.Logf("ClaimCreatorTradingFee instruction data length: %d bytes", len(data))
	t.Logf("Discriminator: %x", data[:8])
}

func TestWithdrawLeftoverInstruction(t *testing.T) {
	// Test data - use valid base58 public keys
	config := solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	pool := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")
	baseVault := solana.MustPublicKeyFromBase58("5e9i1qXq8B9vqvCPrYZhgVTYDPuZyqTZjGHGkwG9i3Vn")
	baseMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	leftoverReceiver := solana.MustPublicKeyFromBase58("GJBKevjCRGwBgj42Wz2Z7sYLHQpB9DYQbLgfYP6tM9vK")
	leftoverReceiverTokenAccount := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	creator := solana.MustPublicKeyFromBase58("AhKa6KGvp5t9KXsqr3FLzjQ7cFQk8YnQWsSmXdJQc39e")

	instruction := WithdrawLeftoverInstruction(
		config, pool, baseVault, baseMint, leftoverReceiver,
		leftoverReceiverTokenAccount, creator,
	)

	// Verify instruction structure
	programID := instruction.ProgramID()
	if !programID.Equals(GetDynamicBondingCurveProgramID()) {
		t.Errorf("Expected program ID %s, got %s", GetDynamicBondingCurveProgramID(), programID)
	}

	// Verify instruction data is not empty and starts with correct discriminator
	data, err := instruction.Data()
	if err != nil {
		t.Fatalf("Failed to get instruction data: %v", err)
	}
	if len(data) < 8 {
		t.Errorf("Instruction data too short: %d bytes", len(data))
	}

	// Check that discriminator matches
	expectedDiscriminator := CreatorWithdrawSurplusDiscriminator
	for i := 0; i < 8; i++ {
		if data[i] != expectedDiscriminator[i] {
			t.Errorf("Discriminator byte %d: expected %02x, got %02x", 
				i, expectedDiscriminator[i], data[i])
		}
	}

	t.Logf("WithdrawLeftover instruction data length: %d bytes", len(data))
	t.Logf("Discriminator: %x", data[:8])
}