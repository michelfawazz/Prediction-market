package dfns

import (
	"os"
)

// Config holds DFNS configuration
type Config struct {
	BaseURL        string // https://api.dfns.io or https://api.dfns.ninja (testnet)
	AppID          string
	OrgID          string
	ServiceKeyPath string // Path to service account private key file
	WebhookSecret  string // Secret for webhook signature verification
}

// LoadConfigFromEnv loads DFNS configuration from environment variables
func LoadConfigFromEnv() Config {
	return Config{
		BaseURL:        getEnvOrDefault("DFNS_API_URL", "https://api.dfns.io"),
		AppID:          os.Getenv("DFNS_APP_ID"),
		OrgID:          os.Getenv("DFNS_ORG_ID"),
		ServiceKeyPath: os.Getenv("DFNS_SERVICE_KEY_PATH"),
		WebhookSecret:  os.Getenv("DFNS_WEBHOOK_SECRET"),
	}
}

// IsConfigured returns true if DFNS is properly configured
func (c Config) IsConfigured() bool {
	return c.AppID != "" && c.OrgID != ""
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
