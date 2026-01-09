package dfns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

// Client is the DFNS API client
type Client struct {
	config     Config
	httpClient *http.Client
	dfnsClient *http.Client
}

// NewClient creates a new DFNS client
func NewClient(config Config) (*Client, error) {
	client := &Client{
		config: config,
	}

	// Load private key content
	privateKey := config.PrivateKey
	if privateKey == "" && config.PrivateKeyPath != "" {
		keyData, err := os.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		privateKey = string(keyData)
	}

	// Create the DFNS signer (no error returned)
	signer := credentials.NewAsymmetricKeySigner(&credentials.AsymmetricKeySignerConfig{
		PrivateKey: privateKey,
		CredID:     config.CredentialID,
	})

	// Create the DFNS API client options
	// Note: OrgID and AuthToken are pointers in the config
	orgID := config.OrgID
	authToken := config.ServiceAccountToken

	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		OrgID:     &orgID,
		AuthToken: &authToken,
		BaseURL:   config.BaseURL,
	}, signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create API options: %w", err)
	}

	// Create the DFNS HTTP client (handles signing automatically)
	client.dfnsClient = api.CreateDfnsAPIClient(apiOptions)
	client.httpClient = &http.Client{}

	return client, nil
}

// doRequest performs an authenticated request to the DFNS API
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	url := c.config.BaseURL + path

	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use the DFNS client which handles signing automatically
	resp, err := c.dfnsClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// APIError represents a DFNS API error
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e APIError) Error() string {
	return fmt.Sprintf("DFNS API error (%d): %s - %s", e.StatusCode, e.Message, e.Details)
}
