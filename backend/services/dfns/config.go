package dfns

import (
	"os"
)

// Config holds DFNS configuration
type Config struct {
	BaseURL             string // https://api.dfns.io or https://api.dfns.ninja (testnet)
	OrgID               string // Organization ID from DFNS dashboard
	ServiceAccountToken string // Service account authentication token
	CredentialID        string // Credential ID for signing (from DFNS dashboard)
	PrivateKey          string // Private key PEM content (for signing)
	PrivateKeyPath      string // Path to service account private key file (for signing)
	WebhookSecret       string // Secret for webhook signature verification
}

// LoadConfigFromEnv loads DFNS configuration from environment variables
func LoadConfigFromEnv() Config {
	return Config{
		BaseURL:             getEnvOrDefault("DFNS_API_URL", "https://api.dfns.io"),
		OrgID:               os.Getenv("DFNS_ORG_ID"),
		ServiceAccountToken: os.Getenv("DFNS_SERVICE_ACCOUNT_TOKEN"),
		CredentialID:        os.Getenv("DFNS_CREDENTIAL_ID"),
		PrivateKey:          os.Getenv("DFNS_PRIVATE_KEY"),
		PrivateKeyPath:      os.Getenv("DFNS_PRIVATE_KEY_PATH"),
		WebhookSecret:       os.Getenv("DFNS_WEBHOOK_SECRET"),
	}
}

// IsConfigured returns true if DFNS is properly configured
func (c Config) IsConfigured() bool {
	return c.ServiceAccountToken != ""
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
