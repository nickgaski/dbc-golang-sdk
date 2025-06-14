package dbc

import (
	"github.com/gagliardetto/solana-go"
	"lukechampine.com/uint128"
)

// Base Fee Configuration
type BaseFeeConfig struct {
	CliffFeeNumerator uint64
	PeriodFrequency   uint64
	ReductionFactor   uint64
	NumberOfPeriod    uint16
	FeeSchedulerMode  uint8
	Padding0          [5]uint8
}

// Dynamic Fee Configuration
type DynamicFeeConfig struct {
	Initialized              uint8
	Padding                  [7]uint8
	MaxVolatilityAccumulator uint32
	VariableFeeControl       uint32
	BinStep                  uint16
	FilterPeriod             uint16
	DecayPeriod              uint16
	ReductionFactor          uint16
	Padding2                 [8]uint8
	BinStepU128              uint128.Uint128
}

// Pool Fees Configuration
type PoolFeesConfig struct {
	BaseFee            BaseFeeConfig
	DynamicFee         DynamicFeeConfig
	Padding0           [5]uint64
	Padding1           [6]uint8
	ProtocolFeePercent uint8
	ReferralFeePercent uint8
}

// Liquidity Distribution Configuration
type LiquidityDistributionConfig struct {
	SqrtPrice uint128.Uint128
	Liquidity uint128.Uint128
}

// Locked Vesting Configuration
type LockedVestingConfig struct {
	AmountPerPeriod                uint64
	CliffDurationFromMigrationTime uint64
	Frequency                      uint64
	NumberOfPeriod                 uint64
	CliffUnlockAmount              uint64
	Padding                        uint64
}

// Pool Configuration
type PoolConfig struct {
	QuoteMint                   solana.PublicKey
	FeeClaimer                  solana.PublicKey
	LeftoverReceiver            solana.PublicKey
	PoolFees                    PoolFeesConfig
	CollectFeeMode              uint8
	MigrationOption             uint8
	ActivationType              uint8
	TokenDecimal                uint8
	Version                     uint8
	TokenType                   uint8
	QuoteTokenFlag              uint8
	PartnerLockedLpPercentage   uint8
	PartnerLpPercentage         uint8
	CreatorLockedLpPercentage   uint8
	CreatorLpPercentage         uint8
	MigrationFeeOption          uint8
	FixedTokenSupplyFlag        uint8
	CreatorTradingFeePercentage uint8
	Padding0                    [2]uint8
	Padding1                    [8]uint8
	SwapBaseAmount              uint64
	MigrationQuoteThreshold     uint64
	MigrationBaseThreshold      uint64
	MigrationSqrtPrice          uint128.Uint128
	LockedVestingConfig         LockedVestingConfig
	PreMigrationTokenSupply     uint64
	PostMigrationTokenSupply    uint64
	Padding2                    [2]uint128.Uint128
	SqrtStartPrice              uint128.Uint128
	Curve                       [20]LiquidityDistributionConfig
}

// Volatility Tracker
type VolatilityTracker struct {
	LastUpdateTimestamp   uint64
	Padding               [8]uint8
	SqrtPriceReference    uint128.Uint128
	VolatilityAccumulator uint128.Uint128
	VolatilityReference   uint128.Uint128
}

// Pool Metrics
type PoolMetrics struct {
	TotalProtocolBaseFee  uint64
	TotalProtocolQuoteFee uint64
	TotalTradingBaseFee   uint64
	TotalTradingQuoteFee  uint64
}

// Virtual Pool (Main Pool State)
type VirtualPool struct {
	VolatilityTracker          VolatilityTracker
	Config                     solana.PublicKey
	Creator                    solana.PublicKey
	BaseMint                   solana.PublicKey
	BaseVault                  solana.PublicKey
	QuoteVault                 solana.PublicKey
	BaseReserve                uint64
	QuoteReserve               uint64
	ProtocolBaseFee            uint64
	ProtocolQuoteFee           uint64
	PartnerBaseFee             uint64
	PartnerQuoteFee            uint64
	SqrtPrice                  uint128.Uint128
	ActivationPoint            uint64
	PoolType                   uint8
	IsMigrated                 uint8
	IsPartnerWithdrawSurplus   uint8
	IsProtocolWithdrawSurplus  uint8
	MigrationProgress          uint8
	IsWithdrawLeftover         uint8
	IsCreatorWithdrawSurplus   uint8
	MigrationFeeWithdrawStatus uint8
	Metrics                    PoolMetrics
	FinishCurveTimestamp       uint64
	CreatorBaseFee             uint64
	CreatorQuoteFee            uint64
	Padding1                   [7]uint64
}

// Pool Fee Metrics
type PoolFeeMetrics struct {
	Current struct {
		PartnerBaseFee  uint64
		PartnerQuoteFee uint64
		CreatorBaseFee  uint64
		CreatorQuoteFee uint64
	}
	Total struct {
		TotalTradingBaseFee  uint64
		TotalTradingQuoteFee uint64
	}
}

// Partner Metadata
type PartnerMetadata struct {
	PartnerName   [32]byte
	PartnerWallet solana.PublicKey
	IsPartner     bool
	Padding       [7]uint8
}

// Virtual Pool Metadata
type VirtualPoolMetadata struct {
	Creator     solana.PublicKey
	PoolCreator solana.PublicKey
	TokenName   [32]byte
	TokenSymbol [16]byte
	TokenUri    [200]byte
	Bump        uint8
	Padding     [7]uint8
}

// Swap Quote Result
type SwapQuoteResult struct {
	AmountOut          uint64
	FeeAmount          uint64
	ProtocolFeeAmount  uint64
	SwapDirection      SwapDirection
	PriceImpactPercent float64
}

// Swap Direction enum
type SwapDirection uint8

const (
	SwapDirectionBaseToQuote SwapDirection = 0
	SwapDirectionQuoteToBase SwapDirection = 1
)

// Activation Type enum
type ActivationType uint8

const (
	ActivationTypeTimestamp ActivationType = 0
	ActivationTypeSlot      ActivationType = 1
)

// Token Type enum
type TokenType uint8

const (
	TokenTypeSPL     TokenType = 0
	TokenTypeToken22 TokenType = 1
)

// Migration Option enum
type MigrationOption uint8

const (
	MigrationOptionNone        MigrationOption = 0
	MigrationOptionMeteoraDamm MigrationOption = 1
	MigrationOptionDammV2      MigrationOption = 2
)

// Pool Type enum
type PoolType uint8

const (
	PoolTypePermissionless PoolType = 0
	PoolTypePermissioned   PoolType = 1
)

// Fee Scheduler Mode enum
type FeeSchedulerMode uint8

const (
	FeeSchedulerModeLinear      FeeSchedulerMode = 0
	FeeSchedulerModeExponential FeeSchedulerMode = 1
)

// Legacy Instruction Discriminators (8 bytes each) - kept for backward compatibility
// These will be replaced by dynamic calculation in functions
var (
	// Pool instructions - use the values that are known to work from the original program
	InitializeVirtualPoolWithSplTokenDiscriminator    = [8]byte{140, 85, 215, 176, 102, 54, 104, 79}
	InitializeVirtualPoolWithToken2022Discriminator   = [8]byte{79, 112, 232, 52, 124, 87, 72, 49}
	SwapDiscriminator                                  = [8]byte{248, 198, 158, 145, 225, 117, 135, 200}
	
	// Configuration instructions
	CreateConfigDiscriminator                         = [8]byte{36, 131, 137, 123, 130, 134, 44, 129}
	CreateVirtualPoolMetadataDiscriminator            = [8]byte{162, 144, 170, 26, 74, 187, 245, 154}
	
	// Creator instructions
	ClaimCreatorTradingFeeDiscriminator               = [8]byte{183, 71, 178, 17, 130, 11, 10, 246}
	TransferPoolCreatorDiscriminator                  = [8]byte{142, 226, 112, 152, 103, 58, 173, 203}
	CreatorWithdrawSurplusDiscriminator               = [8]byte{198, 99, 50, 133, 106, 175, 125, 156}
	WithdrawLeftoverDiscriminator                     = [8]byte{136, 45, 91, 69, 71, 11, 141, 135}
	
	// Partner instructions
	CreatePartnerMetadataDiscriminator                = [8]byte{201, 164, 255, 178, 111, 100, 22, 144}
	ClaimTradingFeeDiscriminator                      = [8]byte{186, 170, 202, 179, 189, 49, 101, 29}
	PartnerWithdrawSurplusDiscriminator               = [8]byte{158, 131, 230, 192, 192, 29, 28, 175}
	
	// Migration instructions
	MigrateMeteoraDammDiscriminator                   = [8]byte{38, 199, 64, 184, 144, 199, 138, 70}
	MigrationDammV2Discriminator                      = [8]byte{173, 203, 49, 206, 245, 18, 213, 42}
	MigrateMeteoraDammClaimLpTokenDiscriminator       = [8]byte{222, 125, 205, 156, 157, 76, 235, 238}
	MigrateMeteoraDammLockLpTokenDiscriminator        = [8]byte{67, 110, 188, 201, 79, 27, 79, 162}
	
	// Admin instructions
	ClaimProtocolFeeDiscriminator                     = [8]byte{144, 9, 74, 205, 173, 7, 132, 134}
	CreateClaimFeeOperatorDiscriminator               = [8]byte{210, 125, 117, 54, 33, 188, 192, 232}
	CloseClaimFeeOperatorDiscriminator                = [8]byte{75, 50, 190, 250, 157, 177, 139, 207}
)