package dbc

import (
	"github.com/gagliardetto/solana-go"
)

// Program constants
const (
	// Fee rate basis points (10000 = 100%)
	MaxFeeRate = 10000
	
	// Default values
	DefaultCommitment = "confirmed"
	DefaultTimeout    = 30 // seconds
	
	// Account discriminators
	ConfigAccountDiscriminator = "config"
	PoolAccountDiscriminator   = "pool"
	VaultAccountDiscriminator  = "vault"
)

// Account size constants
const (
	ConfigAccountSize         = 8 + 32 + 8 + 8 + 8 + 8 + 8 + 8 + 32 + 1 // 149 bytes
	PoolAccountSize           = 8 + 32 + 32 + 32 + 32 + 32 + 32 + 32 + 8 + 8 + 8 + 8 + 8 + 1 + 1 // 277 bytes
	MigrationMetadataSize     = 8 + 32 + 32 + 32 + 32 + 32 + 1 + 1 + 8 // 186 bytes
	LockerAccountSize         = 8 + 32 + 8 + 1 + 4 + 16 * 50 + 1 // Max 50 vesting periods
)

// Swap types
const (
	SwapTypeBuy  uint8 = 0
	SwapTypeSell uint8 = 1
)

// Migration versions
const (
	MigrationV1 uint8 = 1
	MigrationV2 uint8 = 2
)

// Error types
type DBCError struct {
	Code    uint32
	Message string
}

func (e DBCError) Error() string {
	return e.Message
}

// Common DBC errors
var (
	ErrInsufficientBalance = DBCError{Code: 6000, Message: "Insufficient balance"}
	ErrSlippageExceeded   = DBCError{Code: 6001, Message: "Slippage tolerance exceeded"}
	ErrPoolNotFound       = DBCError{Code: 6002, Message: "Pool not found"}
	ErrInvalidSwapType    = DBCError{Code: 6003, Message: "Invalid swap type"}
	ErrMaxCapExceeded     = DBCError{Code: 6004, Message: "Maximum cap exceeded"}
	ErrPoolAlreadyMigrated = DBCError{Code: 6005, Message: "Pool already migrated"}
	ErrInvalidMigrationState = DBCError{Code: 6006, Message: "Invalid migration state"}
)

// Config represents the DBC program configuration
type Config struct {
	Discriminator        [8]byte
	Admin               solana.PublicKey
	TradingFeeRate      uint64
	CreatorFeeRate      uint64
	PartnerFeeRate      uint64
	MaxSwapCap          uint64
	MinBaseMintPrice    uint64
	MaxBaseMintPrice    uint64
	PlatformFeeRecipient solana.PublicKey
	Bump                uint8
}

// Detailed pool structure with all fields
type PoolDetailed struct {
	Discriminator          [8]byte
	Config                solana.PublicKey
	BaseMint              solana.PublicKey
	QuoteMint             solana.PublicKey
	Creator               solana.PublicKey
	Partner               solana.PublicKey
	BaseVault             solana.PublicKey
	QuoteVault            solana.PublicKey
	CurrentSupply         uint64
	ReserveRatio          uint64
	MaxBuyCapAmount       uint64
	TotalBaseSwapped      uint64
	TotalQuoteSwapped     uint64
	CreatorFeeEarned      uint64
	PartnerFeeEarned      uint64
	PlatformFeeEarned     uint64
	LastUpdateSlot        uint64
	IsMigrated            uint8
	MigrationVersion      uint8
	Bump                  uint8
}

// SwapEvent represents a swap event log
type SwapEvent struct {
	Pool         solana.PublicKey
	User         solana.PublicKey
	SwapType     uint8
	AmountIn     uint64
	AmountOut    uint64
	Fee          uint64
	PriceImpact  uint64 // In basis points
	Timestamp    int64
}

// MigrationMetadata represents migration metadata account
type MigrationMetadata struct {
	Discriminator             [8]byte
	Pool                     solana.PublicKey
	BaseMint                 solana.PublicKey
	QuoteMint                solana.PublicKey
	Creator                  solana.PublicKey
	Partner                  solana.PublicKey
	MigrationVersion         uint8
	IsMetadataCreated        bool
	MigrationTimestamp       int64
	Bump                     uint8
}

// LockerAccount represents a token locker account
type LockerAccount struct {
	Discriminator    [8]byte
	BaseMint        solana.PublicKey
	LockDuration    uint64
	HasLockedVesting bool
	VestingScheduleLen uint32
	VestingSchedule  []VestingPeriod
	Bump            uint8
}

// LockEscrow represents a lock escrow account
type LockEscrow struct {
	Discriminator   [8]byte
	Pool           solana.PublicKey
	LPMint         solana.PublicKey
	Beneficiary    solana.PublicKey
	Amount         uint64
	LockDuration   uint64
	LockTimestamp  int64
	IsPartnerLock  bool
	IsReleased     bool
	Bump           uint8
}

// VestingPeriod represents a vesting schedule period
type VestingPeriod struct {
	Timestamp uint64
	Amount    uint64
}

// Utility functions for validation

// IsValidPublicKey checks if a public key is valid
func IsValidPublicKey(key string) bool {
	_, err := solana.PublicKeyFromBase58(key)
	return err == nil
}

// IsValidSwapType checks if swap type is valid
func IsValidSwapType(swapType uint8) bool {
	return swapType == SwapTypeBuy || swapType == SwapTypeSell
}

// IsValidFeeRate checks if fee rate is within valid range
func IsValidFeeRate(rate uint64) bool {
	return rate <= MaxFeeRate
}

// CalculateAccountRent calculates rent for account creation
func CalculateAccountRent(size uint64, lamportsPerByteYear uint64) uint64 {
	// Simplified rent calculation
	// In practice, this would use Solana's rent calculation
	return size * lamportsPerByteYear / 365 / 24 / 3600 * 2 // 2 years of rent
}