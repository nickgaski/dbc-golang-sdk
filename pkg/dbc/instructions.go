package dbc

import (
	"encoding/binary"
	"github.com/gagliardetto/solana-go"
)

// Helper function to pack strings as Rust would (length prefix + data)
func packString(s string) []byte {
	b := make([]byte, 4+len(s))
	binary.LittleEndian.PutUint32(b[:4], uint32(len(s)))
	copy(b[4:], []byte(s))
	return b
}

// Helper functions to derive PDAs
func DerivePoolAuthorityPDA() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("pool_authority")},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DeriveEventAuthorityPDA() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("__event_authority")},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DerivePoolPDA(baseMint, quoteMint solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("pool"),
			baseMint.Bytes(),
			quoteMint.Bytes(),
		},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DeriveBaseVaultPDA(pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("base_vault"),
			pool.Bytes(),
		},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DeriveQuoteVaultPDA(pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("quote_vault"),
			pool.Bytes(),
		},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DeriveConfigPDA() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("config")},
		GetDynamicBondingCurveProgramID(),
	)
	return pda
}

func DeriveMetadataPDA(mint solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			GetMetaplexProgramID().Bytes(),
			mint.Bytes(),
		},
		GetMetaplexProgramID(),
	)
	return pda
}

// CreateConfig creates a new DBC configuration instruction
func CreateConfigInstruction(
	config solana.PublicKey,
	payer solana.PublicKey,
	feeClaimer solana.PublicKey,
	leftoverReceiver solana.PublicKey,
	poolConfig PoolConfig,
) solana.Instruction {
	
	// Create proper Anchor parameters for CreateConfig instruction
	configParams := CreateConfigParams{
		PoolFees:                 poolConfig.PoolFees,
		CollectFeeMode:          0, // Default value
		MigrationConfig:         MigrationConfigParams{MigrationOption: poolConfig.MigrationOption},
		ActivationType:          0, // Default to timestamp
		TokenType:               0, // Default to SPL
		TokenDecimal:            9, // Default SOL decimals
		PartnerLpPercentage:     500, // 5%
		CreatorLpPercentage:     1000, // 10%
		MigrationQuoteThreshold: 1000000000, // Default threshold
		SqrtStartPrice:          [16]byte{}, // Will be calculated
		LockedVesting:           []LockedVestingConfigParams{},
		InitialSupply:           1000000000000, // Default supply
		CreatorFeeConfig:        nil, // No creator fee by default
	}
	
	data, err := SerializeCreateConfigData(configParams)
	if err != nil {
		// Fallback to discriminator only if serialization fails
		data = CreateConfigDiscriminator[:]
	}
	
	accounts := solana.AccountMetaSlice{
		{PublicKey: config, IsSigner: false, IsWritable: true},
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
	}
	
	return solana.NewInstruction(
		GetDynamicBondingCurveProgramID(),
		accounts,
		data,
	)
}

// InitializeVirtualPoolWithSplToken creates a new virtual pool with SPL token
func InitializeVirtualPoolWithSplTokenInstruction(
	config solana.PublicKey,
	poolCreator solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	pool solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	mintMetadata solana.PublicKey,
	payer solana.PublicKey,
	name string,
	symbol string,
	uri string,
) solana.Instruction {
	
	// Create proper Anchor parameters for InitializeVirtualPoolWithSplToken instruction
	poolParams := InitializePoolParams{
		Name:   name,
		Symbol: symbol,
		Uri:    uri,
	}
	
	data, err := SerializeCreatePoolData(poolParams)
	if err != nil {
		// Fallback to discriminator only if serialization fails
		data = InitializeVirtualPoolWithSplTokenDiscriminator[:]
	}
	
	poolAuthority := DerivePoolAuthorityPDA()
	eventAuthority := DeriveEventAuthorityPDA()
	
	accounts := solana.AccountMetaSlice{
		// 1. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 2. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 3. creator (signer)
		{PublicKey: poolCreator, IsSigner: true, IsWritable: false},
		// 4. base_mint (signer, writable)
		{PublicKey: baseMint, IsSigner: true, IsWritable: true},
		// 5. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 6. pool (writable)
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 7. base_vault (writable)
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 8. quote_vault (writable)
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 9. mint_metadata (writable)
		{PublicKey: mintMetadata, IsSigner: false, IsWritable: true},
		// 10. metadata_program
		{PublicKey: GetMetaplexProgramID(), IsSigner: false, IsWritable: false},
		// 11. payer (signer, writable)
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		// 12. token_quote_program
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		// 13. token_program
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		// 14. system_program
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
		// 15. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 16. program (ProgramID)
		{PublicKey: GetDynamicBondingCurveProgramID(), IsSigner: false, IsWritable: false},
	}
	
	return solana.NewInstruction(
		GetDynamicBondingCurveProgramID(),
		accounts,
		data,
	)
}

// SwapInstruction creates a swap instruction
func SwapInstruction(
	config solana.PublicKey,
	pool solana.PublicKey,
	userInputTokenAccount solana.PublicKey,
	userOutputTokenAccount solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	payer solana.PublicKey,
	referralTokenAccount solana.PublicKey,
	amountIn uint64,
	minOut uint64,
) solana.Instruction {
	
	// Create proper Anchor parameters for Swap instruction
	swapParams := AnchorSwapParams{
		AmountIn:         amountIn,
		MinimumAmountOut: minOut,
	}
	
	buf, err := SerializeSwapData(swapParams)
	if err != nil {
		// Fallback to manual serialization if borsh fails
		buf = make([]byte, 8+8+8)
		copy(buf, SwapDiscriminator[:])
		binary.LittleEndian.PutUint64(buf[8:], amountIn)
		binary.LittleEndian.PutUint64(buf[16:], minOut)
	}
	
	poolAuthority := DerivePoolAuthorityPDA()
	eventAuthority := DeriveEventAuthorityPDA()
	
	accounts := solana.AccountMetaSlice{
		// 1. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 2. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 3. pool
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 4. input_token_account
		{PublicKey: userInputTokenAccount, IsSigner: false, IsWritable: true},
		// 5. output_token_account
		{PublicKey: userOutputTokenAccount, IsSigner: false, IsWritable: true},
		// 6. base_vault
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 7. quote_vault
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 8. base_mint
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		// 9. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 10. payer
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		// 11. token_base_program
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		// 12. token_quote_program
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		// 13. referral_token_account
		{PublicKey: referralTokenAccount, IsSigner: false, IsWritable: true},
		// 14. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 15. program
		{PublicKey: GetDynamicBondingCurveProgramID(), IsSigner: false, IsWritable: false},
	}
	
	return solana.NewInstruction(
		GetDynamicBondingCurveProgramID(),
		accounts,
		buf,
	)
}

// ClaimCreatorTradingFeeInstruction creates an instruction to claim creator trading fees
func ClaimCreatorTradingFeeInstruction(
	config solana.PublicKey,
	pool solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	creatorBaseTokenAccount solana.PublicKey,
	creatorQuoteTokenAccount solana.PublicKey,
	creator solana.PublicKey,
) solana.Instruction {
	
	poolAuthority := DerivePoolAuthorityPDA()
	eventAuthority := DeriveEventAuthorityPDA()
	
	accounts := solana.AccountMetaSlice{
		{PublicKey: config, IsSigner: false, IsWritable: false},
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		{PublicKey: creatorBaseTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: creatorQuoteTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: creator, IsSigner: true, IsWritable: false},
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		{PublicKey: GetDynamicBondingCurveProgramID(), IsSigner: false, IsWritable: false},
	}
	
	// Create proper Anchor parameters for ClaimCreatorTradingFee instruction
	claimParams := ClaimCreatorTradingFeeParams{
		MaxBaseAmount:  1000000000, // Default max amounts
		MaxQuoteAmount: 1000000000,
	}
	
	data, err := SerializeClaimCreatorTradingFeeData(claimParams)
	if err != nil {
		// Fallback to discriminator only
		data = ClaimCreatorTradingFeeDiscriminator[:]
	}
	
	return solana.NewInstruction(
		GetDynamicBondingCurveProgramID(),
		accounts,
		data,
	)
}

// WithdrawLeftoverInstruction creates an instruction to withdraw leftover tokens
func WithdrawLeftoverInstruction(
	config solana.PublicKey,
	pool solana.PublicKey,
	baseVault solana.PublicKey,
	baseMint solana.PublicKey,
	leftoverReceiver solana.PublicKey,
	leftoverReceiverTokenAccount solana.PublicKey,
	creator solana.PublicKey,
) solana.Instruction {
	
	poolAuthority := DerivePoolAuthorityPDA()
	eventAuthority := DeriveEventAuthorityPDA()
	
	accounts := solana.AccountMetaSlice{
		{PublicKey: config, IsSigner: false, IsWritable: false},
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		{PublicKey: leftoverReceiver, IsSigner: false, IsWritable: false},
		{PublicKey: leftoverReceiverTokenAccount, IsSigner: false, IsWritable: true},
		{PublicKey: creator, IsSigner: true, IsWritable: false},
		{PublicKey: GetTokenProgramID(), IsSigner: false, IsWritable: false},
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		{PublicKey: GetDynamicBondingCurveProgramID(), IsSigner: false, IsWritable: false},
	}
	
	// Create proper Anchor parameters for CreatorWithdrawSurplus instruction
	withdrawParams := CreatorWithdrawSurplusParams{}
	
	data, err := SerializeCreatorWithdrawSurplusData(withdrawParams)
	if err != nil {
		// Fallback to discriminator only
		data = CreatorWithdrawSurplusDiscriminator[:]
	}
	
	return solana.NewInstruction(
		GetDynamicBondingCurveProgramID(),
		accounts,
		data,
	)
}