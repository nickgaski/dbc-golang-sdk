package dbc

import (
	"testing"
)

func TestCalculateAnchorDiscriminator(t *testing.T) {
	tests := []struct {
		instructionName string
		expectedResult  [8]byte // Use actual discriminators from the working implementation
	}{
		{
			instructionName: "create_config",
			expectedResult:  [8]byte{0xc9, 0xcf, 0xf3, 0x72, 0x4b, 0x6f, 0x2f, 0xbd}, // Our calculated value
		},
		{
			instructionName: "initialize_virtual_pool_with_spl_token",
			expectedResult:  [8]byte{0x8c, 0x55, 0xd7, 0xb0, 0x66, 0x36, 0x68, 0x4f}, // Our calculated value
		},
		{
			instructionName: "swap",
			expectedResult:  [8]byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8}, // From meteora_types.go
		},
		{
			instructionName: "claim_creator_trading_fee",
			expectedResult:  [8]byte{0x52, 0xdc, 0xfa, 0xbd, 0x03, 0x55, 0x6b, 0x2d}, // Our calculated value
		},
		{
			instructionName: "creator_withdraw_surplus",
			expectedResult:  [8]byte{0xa5, 0x03, 0x89, 0x07, 0x1c, 0x86, 0x4c, 0x50}, // Our calculated value
		},
	}

	for _, tt := range tests {
		t.Run(tt.instructionName, func(t *testing.T) {
			discriminator := CalculateAnchorDiscriminator(tt.instructionName)
			
			// Check that we get 8 bytes
			if len(discriminator) != 8 {
				t.Errorf("Expected discriminator length 8, got %d", len(discriminator))
			}
			
			// Check all 8 bytes match expected result
			for i := 0; i < 8; i++ {
				if discriminator[i] != tt.expectedResult[i] {
					t.Errorf("Discriminator mismatch for %s at byte %d: expected %02x, got %02x", 
						tt.instructionName, i, tt.expectedResult[i], discriminator[i])
				}
			}
			
			t.Logf("Discriminator for %s: %x", tt.instructionName, discriminator)
		})
	}
}

func TestGetCorrectDiscriminators(t *testing.T) {
	discriminators := GetCorrectDiscriminators()
	
	// Verify all expected discriminators are present
	expectedInstructions := []string{
		"create_config",
		"initialize_virtual_pool_with_spl_token",
		"swap",
		"claim_creator_trading_fee",
		"creator_withdraw_surplus",
	}
	
	for _, instruction := range expectedInstructions {
		if _, exists := discriminators[instruction]; !exists {
			t.Errorf("Missing discriminator for instruction: %s", instruction)
		}
	}
}