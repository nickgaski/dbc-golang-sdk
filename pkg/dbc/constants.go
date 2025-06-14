package dbc

import (
	"math/big"
	"github.com/gagliardetto/solana-go"
)

// Program IDs
const (
	// Meteora Dynamic Bonding Curve Program
	DynamicBondingCurveProgramID = "dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN"
	
	// Migration Programs
	DAMMv1ProgramID = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"
	DAMMv2ProgramID = "cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG"
	VaultProgramID  = "24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi"
	LockerProgramID = "LocpQgucEQHbqNABEYvBvwoxCPsSbG91A1QaQhQQqjn"
	
	// System Programs
	MetaplexProgramID = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
	TokenProgramID    = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	Token2022ProgramID = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
	
	// Native Mint
	NativeMint = "So11111111111111111111111111111111111111112"
	
	// Base Address for PDA derivation
	BaseAddress = "HWzXGcGHy4tcpYfaRDCyLNzXqBTv3E6BttpCH2vJxArv"
)

// Mathematical constants
const (
	OFFSET     = 64
	RESOLUTION = 64
	
	// Fee constants
	FEE_DENOMINATOR     = 1_000_000_000
	MAX_FEE_BPS         = 9900  // 99%
	MIN_FEE_BPS         = 1     // 0.0001%
	MIN_FEE_NUMERATOR   = 100_000     // 0.0001%
	MAX_FEE_NUMERATOR   = 990_000_000 // 99%
	BASIS_POINT_MAX     = 10000
	
	// Curve constants
	MAX_CURVE_POINT = 16
	
	// Percentage constants
	PARTNER_SURPLUS_SHARE             = 80  // 80%
	SWAP_BUFFER_PERCENTAGE            = 25  // 25%
	MAX_SWALLOW_PERCENTAGE            = 20  // 20%
	MAX_MIGRATION_FEE_PERCENTAGE      = 50  // 50%
	MAX_CREATOR_MIGRATION_FEE_PERCENTAGE = 100 // 100%
	
	// Rate limiter constants
	MAX_RATE_LIMITER_DURATION_IN_SECONDS = 43200  // 12 hours
	MAX_RATE_LIMITER_DURATION_IN_SLOTS   = 108000 // 12 hours
	
	// Time constants
	SLOT_DURATION      = 400
	TIMESTAMP_DURATION = 1000
	
	// Dynamic fee defaults
	DYNAMIC_FEE_FILTER_PERIOD_DEFAULT    = 10
	DYNAMIC_FEE_DECAY_PERIOD_DEFAULT     = 120
	DYNAMIC_FEE_REDUCTION_FACTOR_DEFAULT = 5000 // 50%
	BIN_STEP_BPS_DEFAULT                 = 1
	MAX_PRICE_CHANGE_BPS_DEFAULT         = 1500 // 15%
)

// Big number constants
var (
	U64_MAX          = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	MIN_SQRT_PRICE   = big.NewInt(4295048016)
	MAX_SQRT_PRICE   = new(big.Int)
	ONE_Q64          = new(big.Int).Lsh(big.NewInt(1), RESOLUTION)
	BIN_STEP_BPS_U128_DEFAULT = big.NewInt(1844674407370955) // bin_step << 64 / BASIS_POINT_MAX
)

func init() {
	// MAX_SQRT_PRICE = 79226673521066979257578248091
	MAX_SQRT_PRICE.SetString("79226673521066979257578248091", 10)
}

// DAMM V1 Migration Fee Addresses
var DAMMV1MigrationFeeAddresses = []string{
	"8f848CEy8eY6PhJ3VcemtBDzPPSD4Vq7aJczLZ3o8MmX", // FixedBps25
	"HBxB8Lf14Yj8pqeJ8C4qDb5ryHL7xwpuykz31BLNYr7S", // FixedBps30
	"7v5vBdUQHTNeqk1HnduiXcgbvCyVEZ612HLmYkQoAkik", // FixedBps100
	"EkvP7d5yKxovj884d2DwmBQbrHUWRLGK6bympzrkXGja", // FixedBps200
	"9EZYAJrcqNWNQzP2trzZesP7XKMHA1jEomHzbRsdX8R2", // FixedBps400
	"8cdKo87jZU2R12KY1BUjjRPwyjgdNjLGqSGQyrDshhud", // FixedBps600
}

// DAMM V2 Migration Fee Addresses
var DAMMV2MigrationFeeAddresses = []string{
	"7F6dnUcRuyM2TwR8myT1dYypFXpPSxqwKNSFNkxyNESd", // FixedBps25
	"2nHK1kju6XjphBLbNxpM5XRGFj7p9U8vvNzyZiha1z6k", // FixedBps30
	"Hv8Lmzmnju6m7kcokVKvwqz7QPmdX9XfKjJsXz8RXcjp", // FixedBps100
	"2c4cYd4reUYVRAB9kUUkrq55VPyy2FNQ3FDL4o12JXmq", // FixedBps200
	"AkmQWebAwFvWk55wBoCr5D62C6VVDTzi84NJuD9H7cFD", // FixedBps400
	"DbCRBj8McvPYHJG1ukj8RE15h2dCNUdTAESG49XpQ44u", // FixedBps600
}

// Helper functions to get program IDs as PublicKey
func GetDynamicBondingCurveProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(DynamicBondingCurveProgramID)
}

func GetMetaplexProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(MetaplexProgramID)
}

func GetTokenProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(TokenProgramID)
}

func GetToken2022ProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(Token2022ProgramID)
}

func GetNativeMint() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(NativeMint)
}

func GetBaseAddress() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(BaseAddress)
}