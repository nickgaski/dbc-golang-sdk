package dbc

import (
	"crypto/sha256"
	"fmt"
)

// CalculateAnchorDiscriminator calculates the 8-byte discriminator for an Anchor instruction
// It uses SHA256("global:[instruction_name]")[..8] as per Anchor framework specification
func CalculateAnchorDiscriminator(instructionName string) [8]byte {
	input := fmt.Sprintf("global:%s", instructionName)
	hash := sha256.Sum256([]byte(input))
	
	var discriminator [8]byte
	copy(discriminator[:], hash[:8])
	return discriminator
}

// GetCorrectDiscriminators returns the correctly calculated discriminators
func GetCorrectDiscriminators() map[string][8]byte {
	return map[string][8]byte{
		"create_config":                            CalculateAnchorDiscriminator("create_config"),
		"initialize_virtual_pool_with_spl_token":   CalculateAnchorDiscriminator("initialize_virtual_pool_with_spl_token"),
		"initialize_virtual_pool_with_token2022":   CalculateAnchorDiscriminator("initialize_virtual_pool_with_token2022"),
		"swap":                                     CalculateAnchorDiscriminator("swap"),
		"claim_creator_trading_fee":                CalculateAnchorDiscriminator("claim_creator_trading_fee"),
		"creator_withdraw_surplus":                 CalculateAnchorDiscriminator("creator_withdraw_surplus"),
		"claim_trading_fee":                        CalculateAnchorDiscriminator("claim_trading_fee"),
		"partner_withdraw_surplus":                 CalculateAnchorDiscriminator("partner_withdraw_surplus"),
		"migrate_meteora_damm":                     CalculateAnchorDiscriminator("migrate_meteora_damm"),
		"migration_damm_v2":                        CalculateAnchorDiscriminator("migration_damm_v2"),
		"claim_protocol_fee":                       CalculateAnchorDiscriminator("claim_protocol_fee"),
	}
}

// Anchor instruction parameter structures that match the Rust program exactly

// CreateConfigParams represents the parameters for CreateConfig instruction
type CreateConfigParams struct {
	PoolFees                  PoolFeesConfig               `borsh:"pool_fees"`
	CollectFeeMode           uint8                        `borsh:"collect_fee_mode"`
	MigrationConfig          MigrationConfigParams        `borsh:"migration_config"`
	ActivationType           uint8                        `borsh:"activation_type"`
	TokenType                uint8                        `borsh:"token_type"`
	TokenDecimal             uint8                        `borsh:"token_decimal"`
	PartnerLpPercentage      uint64                       `borsh:"partner_lp_percentage"`
	CreatorLpPercentage      uint64                       `borsh:"creator_lp_percentage"`
	MigrationQuoteThreshold  uint64                       `borsh:"migration_quote_threshold"`
	SqrtStartPrice           [16]byte                     `borsh:"sqrt_start_price"` // u128 as [16]byte
	LockedVesting            []LockedVestingConfigParams  `borsh:"locked_vesting"`
	InitialSupply            uint64                       `borsh:"initial_supply"`
	CreatorFeeConfig         *CreatorFeeConfigParams      `borsh:"creator_fee_config"` // Option<T>
}

// MigrationConfigParams represents migration configuration
type MigrationConfigParams struct {
	MigrationOption uint8 `borsh:"migration_option"`
}

// LockedVestingConfigParams represents locked vesting configuration
type LockedVestingConfigParams struct {
	AmountPerPeriod                uint64 `borsh:"amount_per_period"`
	CliffDurationFromMigrationTime uint64 `borsh:"cliff_duration_from_migration_time"`
	Frequency                      uint64 `borsh:"frequency"`
	NumberOfPeriod                 uint64 `borsh:"number_of_period"`
	CliffUnlockAmount              uint64 `borsh:"cliff_unlock_amount"`
}

// CreatorFeeConfigParams represents creator fee configuration
type CreatorFeeConfigParams struct {
	CreatorTradingFeePercentage uint8 `borsh:"creator_trading_fee_percentage"`
}

// InitializePoolParams represents the parameters for InitializeVirtualPoolWithSplToken instruction
type InitializePoolParams struct {
	Name   string `borsh:"name"`
	Symbol string `borsh:"symbol"`
	Uri    string `borsh:"uri"`
}

// AnchorSwapParams represents the parameters for Swap instruction in Anchor format
type AnchorSwapParams struct {
	AmountIn         uint64 `borsh:"amount_in"`
	MinimumAmountOut uint64 `borsh:"minimum_amount_out"`
}

// ClaimCreatorTradingFeeParams represents the parameters for ClaimCreatorTradingFee instruction
type ClaimCreatorTradingFeeParams struct {
	MaxBaseAmount  uint64 `borsh:"max_base_amount"`
	MaxQuoteAmount uint64 `borsh:"max_quote_amount"`
}

// No parameters needed for creator_withdraw_surplus instruction
type CreatorWithdrawSurplusParams struct {
	// Empty struct - Anchor will serialize as empty
}