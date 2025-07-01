
package serverless

import (
	"fmt"
	"time"
)

// ServerlessWorkConfig represents serverless-specific work configuration
type ServerlessWorkConfig struct {
	Provider    string            `json:"provider"`    // aws, gcp, azure
	Function    string            `json:"function"`
	Region      string            `json:"region,omitempty"`
	Credentials CredentialConfig  `json:"credentials,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Runtime     string            `json:"runtime,omitempty"`
	Memory      int               `json:"memory,omitempty"`      // MB
	Async       bool              `json:"async,omitempty"`       // Asynchronous invocation
	Qualifier   string            `json:"qualifier,omitempty"`   // Version or alias
}

// CredentialConfig defines authentication configuration
type CredentialConfig struct {
	Type        string            `json:"type"`         // aws_iam, service_account, managed_identity
	AccessKey   string            `json:"access_key,omitempty"`
	SecretKey   string            `json:"secret_key,omitempty"`
	SessionToken string           `json:"session_token,omitempty"`
	Profile     string            `json:"profile,omitempty"`
	RoleARN     string            `json:"role_arn,omitempty"`
	KeyFile     string            `json:"key_file,omitempty"`
	ProjectID   string            `json:"project_id,omitempty"`
	ClientID    string            `json:"client_id,omitempty"`
	ClientSecret string           `json:"client_secret,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// DefaultServerlessConfig returns a default serverless configuration
func DefaultServerlessConfig() ServerlessWorkConfig {
	return ServerlessWorkConfig{
		Environment: make(map[string]string),
		Timeout:     5 * time.Minute,
		Memory:      128, // 128 MB default
		Async:       false,
		Credentials: CredentialConfig{
			Properties: make(map[string]string),
		},
	}
}

// Validate validates the serverless configuration
func (c *ServerlessWorkConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("serverless provider is required")
	}
	
	if c.Function == "" {
		return fmt.Errorf("function name is required")
	}
	
	// Validate provider-specific requirements
	switch c.Provider {
	case "aws":
		return c.validateAWS()
	case "gcp":
		return c.validateGCP()
	case "azure":
		return c.validateAzure()
	default:
		return fmt.Errorf("unsupported provider: %s", c.Provider)
	}
}

// validateAWS validates AWS-specific configuration
func (c *ServerlessWorkConfig) validateAWS() error {
	if c.Region == "" {
		return fmt.Errorf("AWS region is required")
	}
	
	// Validate credentials if provided
	if c.Credentials.Type != "" {
		switch c.Credentials.Type {
		case "aws_iam":
			if c.Credentials.AccessKey == "" || c.Credentials.SecretKey == "" {
				return fmt.Errorf("AWS IAM credentials require access_key and secret_key")
			}
		case "aws_role":
			if c.Credentials.RoleARN == "" {
				return fmt.Errorf("AWS role credentials require role_arn")
			}
		}
	}
	
	if c.Memory < 128 || c.Memory > 10240 {
		return fmt.Errorf("AWS Lambda memory must be between 128 and 10240 MB")
	}
	
	return nil
}

// validateGCP validates GCP-specific configuration
func (c *ServerlessWorkConfig) validateGCP() error {
	if c.Region == "" {
		return fmt.Errorf("GCP region is required")
	}
	
	// Validate credentials if provided
	if c.Credentials.Type != "" {
		switch c.Credentials.Type {
		case "service_account":
			if c.Credentials.KeyFile == "" && c.Credentials.ClientID == "" {
				return fmt.Errorf("GCP service account credentials require key_file or client_id")
			}
		}
	}
	
	if c.Memory < 128 || c.Memory > 8192 {
		return fmt.Errorf("GCP Cloud Functions memory must be between 128 and 8192 MB")
	}
	
	return nil
}

// validateAzure validates Azure-specific configuration
func (c *ServerlessWorkConfig) validateAzure() error {
	// Validate credentials if provided
	if c.Credentials.Type != "" {
		switch c.Credentials.Type {
		case "managed_identity":
			// No additional validation needed
		case "service_principal":
			if c.Credentials.ClientID == "" || c.Credentials.ClientSecret == "" || c.Credentials.TenantID == "" {
				return fmt.Errorf("Azure service principal credentials require client_id, client_secret, and tenant_id")
			}
		}
	}
	
	if c.Memory < 128 || c.Memory > 1536 {
		return fmt.Errorf("Azure Functions memory must be between 128 and 1536 MB")
	}
	
	return nil
}

// GetProviderSpecificConfig returns provider-specific configuration
func (c *ServerlessWorkConfig) GetProviderSpecificConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	switch c.Provider {
	case "aws":
		config["FunctionName"] = c.Function
		config["Region"] = c.Region
		if c.Qualifier != "" {
			config["Qualifier"] = c.Qualifier
		}
		if c.Async {
			config["InvocationType"] = "Event"
		} else {
			config["InvocationType"] = "RequestResponse"
		}
		
	case "gcp":
		config["Name"] = c.Function
		config["Location"] = c.Region
		if c.Credentials.ProjectID != "" {
			config["ProjectID"] = c.Credentials.ProjectID
		}
		
	case "azure":
		config["FunctionName"] = c.Function
		if c.Region != "" {
			config["Location"] = c.Region
		}
	}
	
	return config
}
