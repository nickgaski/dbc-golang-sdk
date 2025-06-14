package dbc

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// ClientConfig holds configuration options for the DBC client
type ClientConfig struct {
	// RPC configuration
	RPCEndpoint     string
	Commitment      rpc.CommitmentType
	RPCTimeout      time.Duration
	
	// Transaction configuration
	MaxRetries      int
	ConfirmTimeout  time.Duration
	RetryDelay      time.Duration
	
	// Security settings
	SkipPreflight   bool
	MaxSigners     int
	
	// Performance settings
	BatchSize      int
	PoolingEnabled bool
}

// DefaultConfig returns a production-ready default configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		RPCEndpoint:    "https://api.testnet.solana.com",
		Commitment:     rpc.CommitmentConfirmed,
		RPCTimeout:     30 * time.Second,
		MaxRetries:     3,
		ConfirmTimeout: 30 * time.Second,
		RetryDelay:     2 * time.Second,
		SkipPreflight:  false,
		MaxSigners:     10,
		BatchSize:      10,
		PoolingEnabled: true,
	}
}

// TestnetConfig returns configuration optimized for testnet
func TestnetConfig() *ClientConfig {
	config := DefaultConfig()
	config.RPCEndpoint = "https://api.testnet.solana.com"
	config.Commitment = rpc.CommitmentConfirmed
	config.ConfirmTimeout = 45 * time.Second
	config.MaxRetries = 4
	return config
}

// DevnetConfig returns configuration optimized for devnet
func DevnetConfig() *ClientConfig {
	config := DefaultConfig()
	config.RPCEndpoint = "https://api.devnet.solana.com"
	config.Commitment = rpc.CommitmentConfirmed
	config.ConfirmTimeout = 45 * time.Second
	config.MaxRetries = 4
	return config
}


// NewDBCClientWithConfig creates a new DBC client with custom configuration
func NewDBCClientWithConfig(config *ClientConfig, payer solana.PrivateKey) *DBCClient {
	rpcClient := rpc.New(config.RPCEndpoint)
	
	return &DBCClient{
		RPCClient:   rpcClient,
		ProgramID:   ProgramID,
		Payer:       payer,
		Commitment:  config.Commitment,
		Config:      config,
		RPCEndpoint: config.RPCEndpoint,
	}
}

// Validate checks if the configuration is valid
func (c *ClientConfig) Validate() error {
	if c.RPCEndpoint == "" {
		return fmt.Errorf("RPC endpoint cannot be empty")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	if c.ConfirmTimeout <= 0 {
		return fmt.Errorf("confirm timeout must be positive")
	}
	
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}
	
	if c.MaxSigners <= 0 {
		return fmt.Errorf("max signers must be positive")
	}
	
	return nil
}

// Security utilities for production use

// ValidatePrivateKey ensures the private key is valid and secure
func ValidatePrivateKey(privateKey solana.PrivateKey) error {
	// Check if it's a zero key (all bytes are zero)
	zeroKey := make([]byte, 64)
	if bytes.Equal(privateKey[:], zeroKey) {
		return fmt.Errorf("private key cannot be zero")
	}
	
	// Check if it's a known test key (for security)
	testKeys := []string{
		"11111111111111111111111111111111",
		"5J8g5DWV3rB2RD3K9s8tL3Q1B8F4H9vP6X2C7M4N1K9p",
	}
	
	keyStr := privateKey.String()
	for _, testKey := range testKeys {
		if keyStr == testKey {
			return fmt.Errorf("test private key detected - use a secure key in production")
		}
	}
	
	return nil
}

// ValidatePublicKeyInput validates user-provided public keys
func ValidatePublicKeyInput(pubkey solana.PublicKey) error {
	if pubkey.IsZero() {
		return fmt.Errorf("public key cannot be zero")
	}
	
	// Check for system program ID misuse
	if pubkey.Equals(solana.SystemProgramID) {
		return fmt.Errorf("cannot use system program ID as user key")
	}
	
	return nil
}

// Rate limiting for production use
type RateLimiter struct {
	requests chan struct{}
	ticker   *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(chan struct{}, requestsPerSecond),
		ticker:   time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
	}
	
	// Fill initial bucket
	for i := 0; i < requestsPerSecond; i++ {
		rl.requests <- struct{}{}
	}
	
	// Refill bucket periodically
	go func() {
		for range rl.ticker.C {
			select {
			case rl.requests <- struct{}{}:
			default:
				// Bucket is full
			}
		}
	}()
	
	return rl
}

// Wait blocks until a request can be made
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.requests:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.ticker.Stop()
	close(rl.requests)
}