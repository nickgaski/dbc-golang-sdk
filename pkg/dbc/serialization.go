package dbc

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/near/borsh-go"
)

// CreateConfigData represents the data for CreateConfig instruction
type CreateConfigData struct {
	Admin                    solana.PublicKey
	BaseFee                  BaseFeeConfig
	DynamicFee               DynamicFeeConfig
	ProtocolFeePercent       uint8
	ReferralFeePercent       uint8
	DefaultReferralAccount   solana.PublicKey
	MigrationOption          uint8
	PlatformFeeRecipient     solana.PublicKey
}

// CreatePoolData represents the data for CreatePool instruction
type CreatePoolData struct {
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

// SwapData represents the data for Swap instruction
type SwapData struct {
	SwapAmount    uint64
	MinAmountOut  uint64
	SwapType      uint8
}

// ClaimTradingFeeData represents the data for ClaimTradingFee instruction
type ClaimTradingFeeData struct {
	MaxAmountA uint64
	MaxAmountB uint64
}

// WithdrawLeftoverData represents the data for WithdrawLeftover instruction
type WithdrawLeftoverData struct {
	// No additional data needed beyond accounts
}

// SwapQuoteData represents the data for SwapQuote instruction
type SwapQuoteData struct {
	SwapAmount uint64
	SwapType   uint8
}

// SerializeCreateConfigData serializes CreateConfig instruction data using original discriminator
func SerializeCreateConfigData(params CreateConfigParams) ([]byte, error) {
	var buf bytes.Buffer
	
	// Use the original discriminator from meteora_types.go (known to work)
	buf.Write(CreateConfigDiscriminator[:])
	
	// Serialize the parameters using borsh
	serialized, err := borsh.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CreateConfig parameters: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// SerializeCreatePoolData serializes InitializeVirtualPoolWithSplToken instruction data using original discriminator
func SerializeCreatePoolData(params InitializePoolParams) ([]byte, error) {
	var buf bytes.Buffer
	
	// Use the original discriminator from meteora_types.go (known to work)
	buf.Write(InitializeVirtualPoolWithSplTokenDiscriminator[:])
	
	// Serialize the parameters using borsh
	serialized, err := borsh.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize InitializePool parameters: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// SerializeSwapData serializes Swap instruction data using original discriminator
func SerializeSwapData(params AnchorSwapParams) ([]byte, error) {
	var buf bytes.Buffer
	
	// Use the original discriminator from meteora_types.go (known to work)
	buf.Write(SwapDiscriminator[:])
	
	// Serialize the parameters using borsh
	serialized, err := borsh.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Swap parameters: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// SerializeClaimCreatorTradingFeeData serializes ClaimCreatorTradingFee instruction data using original discriminator
func SerializeClaimCreatorTradingFeeData(params ClaimCreatorTradingFeeParams) ([]byte, error) {
	var buf bytes.Buffer
	
	// Use the original discriminator from meteora_types.go (known to work)
	buf.Write(ClaimCreatorTradingFeeDiscriminator[:])
	
	// Serialize the parameters using borsh
	serialized, err := borsh.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ClaimCreatorTradingFee parameters: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// SerializeCreatorWithdrawSurplusData serializes CreatorWithdrawSurplus instruction data using original discriminator
func SerializeCreatorWithdrawSurplusData(params CreatorWithdrawSurplusParams) ([]byte, error) {
	var buf bytes.Buffer
	
	// Use the original discriminator from meteora_types.go (known to work)
	buf.Write(CreatorWithdrawSurplusDiscriminator[:])
	
	// Serialize the parameters using borsh (empty struct)
	serialized, err := borsh.Serialize(params)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CreatorWithdrawSurplus parameters: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// SerializeSwapQuoteData serializes SwapQuote instruction data using borsh
func SerializeSwapQuoteData(data SwapQuoteData) ([]byte, error) {
	var buf bytes.Buffer
	
	// Write discriminator - SwapQuote uses same discriminator as Swap since it's a view function
	buf.Write(SwapDiscriminator[:])
	
	// Serialize the data using borsh
	serialized, err := borsh.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize SwapQuote data: %w", err)
	}
	
	buf.Write(serialized)
	return buf.Bytes(), nil
}

// Helper function to serialize strings with length prefix (Rust String format)
func SerializeString(s string) []byte {
	var buf bytes.Buffer
	
	// Write string length as u32 little endian
	length := uint32(len(s))
	binary.Write(&buf, binary.LittleEndian, length)
	
	// Write string bytes
	buf.WriteString(s)
	
	return buf.Bytes()
}

// Helper function to serialize Vec<u8> with length prefix
func SerializeBytes(data []byte) []byte {
	var buf bytes.Buffer
	
	// Write length as u32 little endian
	length := uint32(len(data))
	binary.Write(&buf, binary.LittleEndian, length)
	
	// Write bytes
	buf.Write(data)
	
	return buf.Bytes()
}

// Helper function to serialize Option<T> (1 byte discriminator + optional data)
func SerializeSome[T any](data T, serializer func(T) []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(1) // Some variant
	buf.Write(serializer(data))
	return buf.Bytes()
}

func SerializeNone() []byte {
	return []byte{0} // None variant
}