package dbc

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gagliardetto/solana-go"
	"github.com/near/borsh-go"
)

// PDA derivation utilities

// GetConfigPDA derives the config PDA for a given admin
func GetConfigPDA(admin solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("config"),
		admin.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetPoolPDA derives the pool PDA for given mints
func GetPoolPDA(baseMint, quoteMint solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("pool"),
		baseMint.Bytes(),
		quoteMint.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetBaseVaultPDA derives the base vault PDA for a pool
func GetBaseVaultPDA(pool solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("base_vault"),
		pool.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetQuoteVaultPDA derives the quote vault PDA for a pool
func GetQuoteVaultPDA(pool solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("quote_vault"),
		pool.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetMigrationMetadataPDA derives the migration metadata PDA
func GetMigrationMetadataPDA(pool solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("migration_metadata"),
		pool.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetLockerPDA derives the locker PDA for a base mint
func GetLockerPDA(baseMint solana.PublicKey, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("locker"),
		baseMint.Bytes(),
	}
	return solana.FindProgramAddress(seeds, programID)
}

// GetLockEscrowPDA derives the lock escrow PDA
func GetLockEscrowPDA(pool solana.PublicKey, lockType string, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		[]byte("lock_escrow"),
		pool.Bytes(),
		[]byte(lockType), // "partner" or "creator"
	}
	return solana.FindProgramAddress(seeds, programID)
}

// Mathematical utilities for bonding curve calculations

// CalculateBuyReturn calculates tokens received for a buy order
func CalculateBuyReturn(supply, reserveBalance, reserveRatio, depositAmount uint64) uint64 {
	if depositAmount == 0 {
		return 0
	}

	// Convert to float64 for calculation
	supplyFloat := float64(supply)
	reserveFloat := float64(reserveBalance)
	ratioFloat := float64(reserveRatio) / 1e9 // Assuming 9 decimal places
	depositFloat := float64(depositAmount)

	// Bancor formula: ΔS = S * ((1 + ΔR/R)^(CW) - 1)
	// Where CW = reserveRatio
	ratio := 1.0 + (depositFloat / reserveFloat)
	result := supplyFloat * (math.Pow(ratio, ratioFloat) - 1.0)

	return uint64(math.Floor(result))
}

// CalculateSellReturn calculates tokens received for a sell order
func CalculateSellReturn(supply, reserveBalance, reserveRatio, sellAmount uint64) uint64 {
	if sellAmount == 0 || sellAmount >= supply {
		return 0
	}

	// Convert to float64 for calculation
	supplyFloat := float64(supply)
	reserveFloat := float64(reserveBalance)
	ratioFloat := float64(reserveRatio) / 1e9
	sellFloat := float64(sellAmount)

	// Bancor formula: ΔR = R * (1 - (1 - ΔS/S)^(1/CW))
	ratio := 1.0 - (sellFloat / supplyFloat)
	result := reserveFloat * (1.0 - math.Pow(ratio, 1.0/ratioFloat))

	return uint64(math.Floor(result))
}

// CalculatePriceImpact calculates the price impact of a trade
func CalculatePriceImpact(amountIn, amountOut uint64, swapType uint8) float64 {
	if amountIn == 0 || amountOut == 0 {
		return 0.0
	}

	amountInFloat := float64(amountIn)
	amountOutFloat := float64(amountOut)

	if swapType == SwapTypeBuy {
		// For buy orders: impact = (1 - (out/in)) * 100
		return (1.0 - (amountOutFloat / amountInFloat)) * 100.0
	} else {
		// For sell orders: impact = ((out/in) - 1) * 100
		return ((amountOutFloat / amountInFloat) - 1.0) * 100.0
	}
}

// CalculateTradingFee calculates the trading fee for a given amount
func CalculateTradingFee(amount, feeRate uint64) uint64 {
	return (amount * feeRate) / 10000
}

// ApplySlippage applies slippage to an amount
func ApplySlippage(amount uint64, slippagePercent float64) uint64 {
	slippageMultiplier := 1.0 - (slippagePercent / 100.0)
	return uint64(float64(amount) * slippageMultiplier)
}

// Data serialization utilities

// SerializeU64 serializes a uint64 to little-endian bytes
func SerializeU64(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)
	return bytes
}

// SerializeU32 serializes a uint32 to little-endian bytes
func SerializeU32(value uint32) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	return bytes
}

// SerializeU8 serializes a uint8 to bytes
func SerializeU8(value uint8) []byte {
	return []byte{value}
}

// DeserializeU64 deserializes little-endian bytes to uint64
func DeserializeU64(data []byte) uint64 {
	if len(data) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(data)
}

// DeserializeU32 deserializes little-endian bytes to uint32
func DeserializeU32(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(data)
}

// Validation utilities

// ValidateAmount checks if an amount is within valid range
func ValidateAmount(amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("amount cannot be zero")
	}
	if amount > math.MaxUint64/2 { // Prevent overflow
		return fmt.Errorf("amount too large")
	}
	return nil
}

// ValidateSlippage checks if slippage is within reasonable range
func ValidateSlippage(slippage float64) error {
	if slippage < 0 {
		return fmt.Errorf("slippage cannot be negative")
	}
	if slippage > 50.0 { // 50% max slippage
		return fmt.Errorf("slippage too high (max 50%%)")
	}
	return nil
}

// ValidatePublicKey validates a Solana public key
func ValidatePublicKey(key solana.PublicKey) error {
	if key.IsZero() {
		return fmt.Errorf("public key cannot be zero")
	}
	return nil
}

// Conversion utilities

// LamportsToSOL converts lamports to SOL
func LamportsToSOL(lamports uint64) float64 {
	return float64(lamports) / 1e9
}

// SOLToLamports converts SOL to lamports
func SOLToLamports(sol float64) uint64 {
	return uint64(sol * 1e9)
}

// TokensToBaseUnit converts human-readable tokens to base units
func TokensToBaseUnit(tokens float64, decimals uint8) uint64 {
	multiplier := math.Pow10(int(decimals))
	return uint64(tokens * multiplier)
}

// BaseUnitToTokens converts base units to human-readable tokens
func BaseUnitToTokens(baseUnit uint64, decimals uint8) float64 {
	divisor := math.Pow10(int(decimals))
	return float64(baseUnit) / divisor
}

// Hash utilities

// HashInstruction creates a hash for instruction identification
func HashInstruction(discriminator []byte, data []byte) [32]byte {
	hasher := sha256.New()
	hasher.Write(discriminator)
	hasher.Write(data)
	var hash [32]byte
	copy(hash[:], hasher.Sum(nil))
	return hash
}

// CreateInstructionDiscriminator creates an 8-byte discriminator from method name
func CreateInstructionDiscriminator(methodName string) [8]byte {
	hash := sha256.Sum256([]byte(fmt.Sprintf("global:%s", methodName)))
	var discriminator [8]byte
	copy(discriminator[:], hash[:8])
	return discriminator
}

// Network utilities

// GetExplorerURL returns the Solana explorer URL for a transaction
func GetExplorerURL(signature string, network string) string {
	if network == "testnet" {
		return fmt.Sprintf("https://explorer.solana.com/tx/%s?cluster=testnet", signature)
	}
	return fmt.Sprintf("https://explorer.solana.com/tx/%s", signature)
}

// GetAccountExplorerURL returns the Solana explorer URL for an account
func GetAccountExplorerURL(account solana.PublicKey, network string) string {
	if network == "testnet" {
		return fmt.Sprintf("https://explorer.solana.com/address/%s?cluster=testnet", account.String())
	}
	return fmt.Sprintf("https://explorer.solana.com/address/%s", account.String())
}

// Deserialization utilities

// DeserializePool deserializes pool data from bytes
func DeserializePool(data []byte, pool *VirtualPool) error {
	return borsh.Deserialize(pool, data)
}

// DeserializeConfig deserializes config data from bytes
func DeserializeConfig(data []byte, config *Config) error {
	return borsh.Deserialize(config, data)
}

// DeserializeMigrationMetadata deserializes migration metadata from bytes
func DeserializeMigrationMetadata(data []byte, metadata *MigrationMetadata) error {
	return borsh.Deserialize(metadata, data)
}

// DeserializeLocker deserializes locker data from bytes
func DeserializeLocker(data []byte, locker *LockerAccount) error {
	return borsh.Deserialize(locker, data)
}

// DeserializeLockEscrow deserializes lock escrow data from bytes
func DeserializeLockEscrow(data []byte, escrow *LockEscrow) error {
	return borsh.Deserialize(escrow, data)
}