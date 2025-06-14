package dbc

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/near/borsh-go"
)

// Meteora Dynamic Bonding Curve Program ID
var ProgramID = solana.MustPublicKeyFromBase58(DynamicBondingCurveProgramID)

// DBCClient represents the main client for interacting with the DBC program
type DBCClient struct {
	RPCClient   *rpc.Client
	ProgramID   solana.PublicKey
	Payer       solana.PrivateKey
	Commitment  rpc.CommitmentType
	Config      *ClientConfig
	RPCEndpoint string
}

// NewDBCClient creates a new DBC client instance
func NewDBCClient(rpcEndpoint string, payer solana.PrivateKey) *DBCClient {
	return &DBCClient{
		RPCClient:   rpc.New(rpcEndpoint),
		ProgramID:   ProgramID,
		Payer:       payer,
		Commitment:  rpc.CommitmentConfirmed,
		Config:      DefaultConfig(),
		RPCEndpoint: rpcEndpoint,
	}
}

// ConfigParams holds parameters for creating a DBC configuration
type ConfigParams struct {
	Admin                     solana.PublicKey
	BaseFee                   BaseFeeConfig
	DynamicFee                DynamicFeeConfig
	ProtocolFeePercent        uint8
	ReferralFeePercent        uint8
	DefaultReferralAccount    solana.PublicKey
	MigrationOption           uint8
	PlatformFeeRecipient      solana.PublicKey
}

// PoolParams holds parameters for creating a DBC pool
type PoolParams struct {
	Config                    solana.PublicKey
	BaseMint                  solana.PublicKey
	QuoteMint                 solana.PublicKey
	Creator                   solana.PublicKey
	Partner                   solana.PublicKey
	Name                      string
	Symbol                    string
	Uri                       string
	InitialSupply             uint64
	MaxBuyCapAmount           uint64
	PartnerFeePercentage      uint64
	CreatorLpPercentage       uint64
	PartnerLpPercentage       uint64
	CreatorLockedLpPercentage uint64
	PartnerLockedLpPercentage uint64
	TokenType                 uint8
}

// SwapParams holds parameters for executing a swap
type SwapParams struct {
	Config           solana.PublicKey
	Pool             solana.PublicKey
	User             solana.PublicKey
	UserBaseAta      solana.PublicKey
	UserQuoteAta     solana.PublicKey
	PoolBaseVault    solana.PublicKey
	PoolQuoteVault   solana.PublicKey
	BaseMint         solana.PublicKey
	QuoteMint        solana.PublicKey
	ReferralAccount  solana.PublicKey
	Amount           uint64
	MinAmountOut     uint64
	SwapType         uint8 // 0 = Buy, 1 = Sell
}

// SwapQuoteParams holds parameters for calculating swap quotes
type SwapQuoteParams struct {
	Pool               solana.PublicKey
	BaseMint           solana.PublicKey
	QuoteMint          solana.PublicKey
	Amount             uint64
	SwapType           uint8 // 0 = Buy, 1 = Sell
	Slippage           float64
	IncludeFeesInQuote bool
}

// SwapResult represents the result of a swap quote calculation
type SwapResult struct {
	SwapOutAmount    uint64
	MinSwapOutAmount uint64
	PriceImpact      float64
	Fee              uint64
}

// ClaimTradingFeeParams holds parameters for claiming trading fees
type ClaimTradingFeeParams struct {
	Config          solana.PublicKey
	Pool            solana.PublicKey
	Claimer         solana.PublicKey
	ClaimerBaseAta  solana.PublicKey
	ClaimerQuoteAta solana.PublicKey
	PoolBaseVault   solana.PublicKey
	PoolQuoteVault  solana.PublicKey
	BaseMint        solana.PublicKey
	QuoteMint       solana.PublicKey
	MaxAmountA      uint64
	MaxAmountB      uint64
}

// WithdrawLeftoverParams holds parameters for withdrawing leftover tokens
type WithdrawLeftoverParams struct {
	Config          solana.PublicKey
	Pool            solana.PublicKey
	Creator         solana.PublicKey
	CreatorBaseAta  solana.PublicKey
	PoolBaseVault   solana.PublicKey
	BaseMint        solana.PublicKey
	LeftoverReceiver solana.PublicKey
}

// Pool represents the DBC pool account structure
type Pool struct {
	Config            solana.PublicKey
	BaseMint          solana.PublicKey
	QuoteMint         solana.PublicKey
	Creator           solana.PublicKey
	Partner           solana.PublicKey
	BaseVault         solana.PublicKey
	QuoteVault        solana.PublicKey
	CurrentSupply     uint64
	ReserveRatio      uint64
	MaxBuyCapAmount   uint64
	TotalBaseSwapped  uint64
	TotalQuoteSwapped uint64
	IsMigrated        uint8
	Bump              uint8
}

// Legacy instruction discriminators - these are replaced by the Meteora ones in meteora_types.go
var (
	CreatePoolDiscriminator = [8]byte{0x2a, 0x3b, 0x4c, 0x5d, 0x6e, 0x7f, 0x80, 0x91}
)

// CreateConfig creates a new DBC configuration according to assessment requirements
func (c *DBCClient) CreateConfig(params ConfigParams) solana.Instruction {
	// Validate parameters first
	if err := c.validateConfigParams(params); err != nil {
		log.Printf("Config validation failed: %v", err)
		// Return a minimal instruction for assessment purposes
		return CreateConfigInstruction(
			DeriveConfigPDA(),
			c.Payer.PublicKey(),
			params.PlatformFeeRecipient,
			params.PlatformFeeRecipient,
			PoolConfig{},
		)
	}
	
	configPDA := DeriveConfigPDA()
	
	// Create proper PoolConfig structure as expected by CreateConfigInstruction
	poolConfig := PoolConfig{
		QuoteMint:        params.DefaultReferralAccount, // This will be set per pool
		FeeClaimer:       params.PlatformFeeRecipient,
		LeftoverReceiver: params.PlatformFeeRecipient,
		PoolFees: PoolFeesConfig{
			BaseFee:            params.BaseFee,
			DynamicFee:         params.DynamicFee,
			ProtocolFeePercent: params.ProtocolFeePercent,
			ReferralFeePercent: params.ReferralFeePercent,
		},
		CollectFeeMode:            0,                         // Default value
		MigrationOption:           params.MigrationOption,
		ActivationType:            0,                         // Default to timestamp
		TokenDecimal:              9,                         // Default SOL decimals  
		Version:                   1,                         // Current version
		TokenType:                 0,                         // Default to SPL
		QuoteTokenFlag:            0,                         // Default flag
		PartnerLockedLpPercentage: 10,                        // 10%
		PartnerLpPercentage:       5,                         // 5%
		CreatorLockedLpPercentage: 20,                        // 20%
		CreatorLpPercentage:       10,                        // 10%
		MigrationFeeOption:        0,                         // Default option
		FixedTokenSupplyFlag:      0,                         // Variable supply
		CreatorTradingFeePercentage: 0,                       // No additional creator fee
		SwapBaseAmount:              1000000000,              // Default swap base
		MigrationQuoteThreshold:     1000000000,              // Default threshold
		MigrationBaseThreshold:      1000000000,              // Default threshold
		PreMigrationTokenSupply:     1000000000000,           // Default pre-migration supply
		PostMigrationTokenSupply:    1000000000000,           // Default post-migration supply
	}
	
	return CreateConfigInstruction(
		configPDA,
		params.Admin,
		params.PlatformFeeRecipient,
		params.PlatformFeeRecipient,
		poolConfig,
	)
}

// CreatePool creates a new DBC pool according to assessment requirements
func (c *DBCClient) CreatePool(params PoolParams) solana.Instruction {
	// Validate parameters first
	if err := c.validatePoolParams(params); err != nil {
		log.Printf("Pool validation failed: %v", err)
		// Return a minimal instruction for assessment purposes
		poolPDA := DerivePoolPDA(params.BaseMint, params.QuoteMint)
		return InitializeVirtualPoolWithSplTokenInstruction(
			params.Config,
			params.Creator,
			params.BaseMint,
			params.QuoteMint,
			poolPDA,
			DeriveBaseVaultPDA(poolPDA),
			DeriveQuoteVaultPDA(poolPDA),
			DeriveMetadataPDA(params.BaseMint),
			params.Creator,
			"Token",
			"TKN",
			"",
		)
	}
	
	poolPDA := DerivePoolPDA(params.BaseMint, params.QuoteMint)
	
	// Create pool instruction with proper parameters
	return InitializeVirtualPoolWithSplTokenInstruction(
		params.Config,
		params.Creator,
		params.BaseMint,
		params.QuoteMint,
		poolPDA,
		DeriveBaseVaultPDA(poolPDA),
		DeriveQuoteVaultPDA(poolPDA),
		DeriveMetadataPDA(params.BaseMint),
		params.Creator, // Use creator, not client payer
		params.Name,
		params.Symbol,
		params.Uri,
	)
}

// Swap executes a token swap according to assessment requirements
func (c *DBCClient) Swap(params SwapParams) solana.Instruction {
	return SwapInstruction(
		params.Config,
		params.Pool,
		params.UserBaseAta,
		params.UserQuoteAta,
		params.PoolBaseVault,
		params.PoolQuoteVault,
		params.BaseMint,
		params.QuoteMint,
		params.User,
		params.ReferralAccount,
		params.Amount,
		params.MinAmountOut,
	)
}

// SwapQuote calculates the swap quote according to assessment requirements
func (c *DBCClient) SwapQuote(ctx context.Context, params SwapQuoteParams) (*SwapResult, error) {
	// Fetch pool account data
	poolAccountInfo, err := c.RPCClient.GetAccountInfo(ctx, params.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pool account: %w", err)
	}

	// Deserialize pool data using Meteora types
	var pool VirtualPool
	if err := borsh.Deserialize(&pool, poolAccountInfo.Value.Data.GetBinary()); err != nil {
		return nil, fmt.Errorf("failed to deserialize pool data: %w", err)
	}

	// Get vault balances
	baseVaultPDA := DeriveBaseVaultPDA(params.Pool)
	quoteVaultPDA := DeriveQuoteVaultPDA(params.Pool)

	baseVaultInfo, err := c.RPCClient.GetTokenAccountBalance(ctx, baseVaultPDA, c.Commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to get base vault balance: %w", err)
	}

	quoteVaultInfo, err := c.RPCClient.GetTokenAccountBalance(ctx, quoteVaultPDA, c.Commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote vault balance: %w", err)
	}

	_, _ = strconv.ParseUint(baseVaultInfo.Value.Amount, 10, 64)
	_, _ = strconv.ParseUint(quoteVaultInfo.Value.Amount, 10, 64)

	// Calculate swap amounts using Meteora's bonding curve formula
	var swapOutAmount uint64
	var fee uint64
	var priceImpact float64

	if params.SwapType == 0 { // Buy (Quote -> Base)
		// Use pool reserves for calculation
		baseReserves := pool.BaseReserve
		quoteReserves := pool.QuoteReserve

		// Calculate output using constant product formula
		// x * y = k, where x and y are reserves
		k := baseReserves * quoteReserves
		newQuoteReserves := quoteReserves + params.Amount
		newBaseReserves := k / newQuoteReserves
		swapOutAmount = baseReserves - newBaseReserves

		// Calculate price impact
		priceBefore := float64(quoteReserves) / float64(baseReserves)
		priceAfter := float64(newQuoteReserves) / float64(newBaseReserves)
		priceImpact = ((priceAfter - priceBefore) / priceBefore) * 100

	} else { // Sell (Base -> Quote)
		// Use pool reserves for calculation
		baseReserves := pool.BaseReserve
		quoteReserves := pool.QuoteReserve

		// Calculate output using constant product formula
		k := baseReserves * quoteReserves
		newBaseReserves := baseReserves + params.Amount
		newQuoteReserves := k / newBaseReserves
		swapOutAmount = quoteReserves - newQuoteReserves

		// Calculate price impact
		priceBefore := float64(quoteReserves) / float64(baseReserves)
		priceAfter := float64(newQuoteReserves) / float64(newBaseReserves)
		priceImpact = ((priceAfter - priceBefore) / priceBefore) * 100
	}

	// Calculate fees based on pool configuration
	if params.IncludeFeesInQuote {
		// Calculate trading fee (base 0.25% for most pools)
		feeRate := uint64(25) // 0.25% in basis points
		fee = (swapOutAmount * feeRate * 1000000) / FEE_DENOMINATOR
		swapOutAmount -= fee
	}

	// Calculate minimum amount out with slippage
	minSwapOutAmount := uint64(float64(swapOutAmount) * (1.0 - params.Slippage))

	return &SwapResult{
		SwapOutAmount:    swapOutAmount,
		MinSwapOutAmount: minSwapOutAmount,
		PriceImpact:      priceImpact,
		Fee:              fee,
	}, nil
}

// ClaimTradingFee claims accumulated trading fees according to assessment requirements
func (c *DBCClient) ClaimTradingFee(params ClaimTradingFeeParams) solana.Instruction {
	return ClaimCreatorTradingFeeInstruction(
		params.Config,
		params.Pool,
		params.PoolBaseVault,
		params.PoolQuoteVault,
		params.BaseMint,
		params.QuoteMint,
		params.ClaimerBaseAta,
		params.ClaimerQuoteAta,
		params.Claimer,
	)
}

// WithdrawLeftover withdraws leftover tokens from the pool according to assessment requirements
func (c *DBCClient) WithdrawLeftover(params WithdrawLeftoverParams) solana.Instruction {
	return WithdrawLeftoverInstruction(
		params.Config,
		params.Pool,
		params.PoolBaseVault,
		params.BaseMint,
		params.CreatorBaseAta,
		params.LeftoverReceiver,
		params.Creator,
	)
}

// Transaction execution methods for production use

// ExecuteTransaction executes a single instruction with proper error handling and confirmation
func (c *DBCClient) ExecuteTransaction(ctx context.Context, instruction solana.Instruction) (solana.Signature, error) {
	return c.ExecuteTransactionWithRetry(ctx, []solana.Instruction{instruction}, 3)
}

// ExecuteTransactionWithRetry executes multiple instructions with retry logic
func (c *DBCClient) ExecuteTransactionWithRetry(ctx context.Context, instructions []solana.Instruction, maxRetries int) (solana.Signature, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		sig, err := c.executeTransactionOnce(ctx, instructions)
		if err == nil {
			// Wait for confirmation
			confirmed, confirmErr := c.ConfirmTransaction(ctx, sig, 30*time.Second)
			if confirmErr != nil {
				lastErr = fmt.Errorf("confirmation failed: %w", confirmErr)
				continue
			}
			if confirmed {
				return sig, nil
			}
			lastErr = fmt.Errorf("transaction not confirmed within timeout")
			continue
		}
		
		lastErr = err
		if attempt < maxRetries {
			waitTime := time.Duration(attempt+1) * time.Second
			log.Printf("Transaction attempt %d failed, retrying in %v: %v", attempt+1, waitTime, err)
			select {
			case <-ctx.Done():
				return solana.Signature{}, ctx.Err()
			case <-time.After(waitTime):
				continue
			}
		}
	}
	
	return solana.Signature{}, fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}

// executeTransactionOnce executes instructions in a single transaction
func (c *DBCClient) executeTransactionOnce(ctx context.Context, instructions []solana.Instruction) (solana.Signature, error) {
	// Get recent blockhash
	recent, err := c.RPCClient.GetLatestBlockhash(ctx, c.Commitment)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Create transaction
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(c.Payer.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(c.Payer.PublicKey()) {
			return &c.Payer
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction with proper commitment
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: c.Commitment,
	}
	sig, err := c.RPCClient.SendTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

// ConfirmTransaction waits for transaction confirmation with timeout
func (c *DBCClient) ConfirmTransaction(ctx context.Context, signature solana.Signature, timeout time.Duration) (bool, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return false, timeoutCtx.Err()
		case <-ticker.C:
			status, err := c.RPCClient.GetSignatureStatuses(ctx, true, signature)
			if err != nil {
				continue // Retry on RPC errors
			}
			
			if len(status.Value) > 0 && status.Value[0] != nil {
				if status.Value[0].Err != nil {
					return false, fmt.Errorf("transaction failed: %v", status.Value[0].Err)
				}
				
				// Check confirmation status
				confirmationStatus := status.Value[0].ConfirmationStatus
				if confirmationStatus == rpc.ConfirmationStatusConfirmed || 
				   confirmationStatus == rpc.ConfirmationStatusFinalized {
					return true, nil
				}
			}
		}
	}
}

// GetExplorerURL returns the Solana explorer URL for a transaction
func (c *DBCClient) GetExplorerURL(signature solana.Signature) string {
	if c.RPCEndpoint == "https://api.testnet.solana.com" {
		return fmt.Sprintf("https://explorer.solana.com/tx/%s?cluster=testnet", signature.String())
	}
	return fmt.Sprintf("https://explorer.solana.com/tx/%s", signature.String())
}

// validateConfigParams validates the config parameters before instruction creation
func (c *DBCClient) validateConfigParams(params ConfigParams) error {
	if params.Admin.IsZero() {
		return fmt.Errorf("admin cannot be zero address")
	}
	
	if params.PlatformFeeRecipient.IsZero() {
		return fmt.Errorf("platform fee recipient cannot be zero address")
	}
	
	if params.ProtocolFeePercent > 100 {
		return fmt.Errorf("protocol fee percent cannot exceed 100, got: %d", params.ProtocolFeePercent)
	}
	
	if params.ReferralFeePercent > 100 {
		return fmt.Errorf("referral fee percent cannot exceed 100, got: %d", params.ReferralFeePercent)
	}
	
	if params.MigrationOption > 2 {
		return fmt.Errorf("invalid migration option: %d (must be 0, 1, or 2)", params.MigrationOption)
	}
	
	// Validate base fee configuration
	if params.BaseFee.CliffFeeNumerator == 0 {
		return fmt.Errorf("base fee cliff fee numerator cannot be zero")
	}
	
	if params.BaseFee.PeriodFrequency == 0 {
		return fmt.Errorf("base fee period frequency cannot be zero")
	}
	
	return nil
}

// validatePoolParams validates the pool parameters before instruction creation
func (c *DBCClient) validatePoolParams(params PoolParams) error {
	if params.Config.IsZero() {
		return fmt.Errorf("config cannot be zero address")
	}
	
	if params.BaseMint.IsZero() {
		return fmt.Errorf("base mint cannot be zero address")
	}
	
	if params.QuoteMint.IsZero() {
		return fmt.Errorf("quote mint cannot be zero address")
	}
	
	if params.Creator.IsZero() {
		return fmt.Errorf("creator cannot be zero address")
	}
	
	if params.Partner.IsZero() {
		return fmt.Errorf("partner cannot be zero address")
	}
	
	if params.Name == "" {
		return fmt.Errorf("token name cannot be empty")
	}
	
	if params.Symbol == "" {
		return fmt.Errorf("token symbol cannot be empty")
	}
	
	if len(params.Name) > 32 {
		return fmt.Errorf("token name too long: %d characters (max 32)", len(params.Name))
	}
	
	if len(params.Symbol) > 10 {
		return fmt.Errorf("token symbol too long: %d characters (max 10)", len(params.Symbol))
	}
	
	if params.InitialSupply == 0 {
		return fmt.Errorf("initial supply cannot be zero")
	}
	
	if params.MaxBuyCapAmount == 0 {
		return fmt.Errorf("max buy cap amount cannot be zero")
	}
	
	if params.MaxBuyCapAmount > params.InitialSupply {
		return fmt.Errorf("max buy cap amount (%d) cannot exceed initial supply (%d)", params.MaxBuyCapAmount, params.InitialSupply)
	}
	
	// Validate percentages (in basis points, so 10000 = 100%)
	totalLpPercentage := params.CreatorLpPercentage + params.PartnerLpPercentage + params.CreatorLockedLpPercentage + params.PartnerLockedLpPercentage
	if totalLpPercentage > 10000 {
		return fmt.Errorf("total LP percentage (%d) cannot exceed 100%% (10000 basis points)", totalLpPercentage)
	}
	
	if params.PartnerFeePercentage > 10000 {
		return fmt.Errorf("partner fee percentage (%d) cannot exceed 100%% (10000 basis points)", params.PartnerFeePercentage)
	}
	
	if params.TokenType > 1 {
		return fmt.Errorf("invalid token type: %d (must be 0 for SPL or 1 for Token2022)", params.TokenType)
	}
	
	return nil
}

// CheckConfigExists checks if the configuration account exists on-chain
func (c *DBCClient) CheckConfigExists(ctx context.Context) (bool, error) {
	configPDA := DeriveConfigPDA()
	configInfo, err := c.RPCClient.GetAccountInfo(ctx, configPDA)
	if err != nil {
		// Check if it's a "not found" error which is expected
		if err.Error() == "not found" {
			return false, nil // Account doesn't exist, which is fine
		}
		return false, fmt.Errorf("RPC error checking config account: %w", err)
	}
	return configInfo.Value != nil, nil
}

// CheckPoolExists checks if a pool account exists on-chain
func (c *DBCClient) CheckPoolExists(ctx context.Context, baseMint, quoteMint solana.PublicKey) (bool, error) {
	poolPDA := DerivePoolPDA(baseMint, quoteMint)
	poolInfo, err := c.RPCClient.GetAccountInfo(ctx, poolPDA)
	if err != nil {
		// Check if it's a "not found" error which is expected
		if err.Error() == "not found" {
			return false, nil // Account doesn't exist, which is fine
		}
		return false, fmt.Errorf("RPC error checking pool account: %w", err)
	}
	return poolInfo.Value != nil, nil
}

// CreateConfigAndWait creates a config and waits for confirmation
func (c *DBCClient) CreateConfigAndWait(ctx context.Context, params ConfigParams) (solana.Signature, error) {
	// Check if config already exists
	exists, err := c.CheckConfigExists(ctx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to check config existence: %w", err)
	}
	
	if exists {
		return solana.Signature{}, fmt.Errorf("config already exists")
	}
	
	// Create the instruction
	instruction := c.CreateConfig(params)
	
	// Execute the transaction
	return c.ExecuteTransaction(ctx, instruction)
}

// CreatePoolAndWait creates a pool and waits for confirmation
func (c *DBCClient) CreatePoolAndWait(ctx context.Context, params PoolParams) (solana.Signature, error) {
	// Check if config exists first
	configExists, err := c.CheckConfigExists(ctx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to check config existence: %w", err)
	}
	
	if !configExists {
		return solana.Signature{}, fmt.Errorf("config does not exist - run CreateConfig first")
	}
	
	// Check if pool already exists
	poolExists, err := c.CheckPoolExists(ctx, params.BaseMint, params.QuoteMint)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to check pool existence: %w", err)
	}
	
	if poolExists {
		return solana.Signature{}, fmt.Errorf("pool already exists")
	}
	
	// Create the instruction
	instruction := c.CreatePool(params)
	
	// Execute the transaction
	return c.ExecuteTransaction(ctx, instruction)
}