
package grpc

import (
	"fmt"
	"time"
)

// GRPCWorkConfig represents gRPC-specific work configuration
type GRPCWorkConfig struct {
	Endpoint    string            `json:"endpoint"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Retry       RetryConfig       `json:"retry,omitempty"`
	TLS         TLSConfig         `json:"tls,omitempty"`
	Compression string            `json:"compression,omitempty"` // gzip, deflate
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RetryConfig defines retry behavior for gRPC calls
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	Jitter       bool          `json:"jitter"`
}

// TLSConfig defines TLS configuration for gRPC connections
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
	CertFile           string `json:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty"`
	ServerName         string `json:"server_name,omitempty"`
}

// DefaultGRPCConfig returns a default gRPC configuration
func DefaultGRPCConfig() GRPCWorkConfig {
	return GRPCWorkConfig{
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
		Retry: RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
		},
		TLS: TLSConfig{
			Enabled: false,
		},
		Metadata: make(map[string]string),
	}
}

// Validate validates the gRPC configuration
func (c *GRPCWorkConfig) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("gRPC endpoint is required")
	}
	
	if c.Method == "" {
		return fmt.Errorf("gRPC method is required")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.Retry.MaxAttempts < 1 {
		return fmt.Errorf("retry max attempts must be at least 1")
	}
	
	if c.Retry.Multiplier <= 0 {
		return fmt.Errorf("retry multiplier must be positive")
	}
	
	if c.TLS.Enabled {
		if c.TLS.CertFile != "" && c.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when cert file is specified")
		}
		if c.TLS.KeyFile != "" && c.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file is required when key file is specified")
		}
	}
	
	return nil
}
