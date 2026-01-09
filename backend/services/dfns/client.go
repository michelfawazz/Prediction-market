package dfns

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is the DFNS API client
type Client struct {
	config     Config
	httpClient *http.Client
	privateKey *rsa.PrivateKey
}

// NewClient creates a new DFNS client
func NewClient(config Config) (*Client, error) {
	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Load private key if path is provided
	if config.ServiceKeyPath != "" {
		key, err := loadPrivateKey(config.ServiceKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		client.privateKey = key
	}

	return client, nil
}

// loadPrivateKey loads an RSA private key from a PEM file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as PKCS8 first (more common for DFNS)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback to PKCS1
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	return rsaKey, nil
}

// signRequest creates a signature for DFNS API authentication
func (c *Client) signRequest(method, path string, body []byte) (string, string, error) {
	if c.privateKey == nil {
		return "", "", fmt.Errorf("no private key configured")
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create the message to sign
	// DFNS uses: method + path + timestamp + body
	message := fmt.Sprintf("%s%s%s%s", method, path, timestamp, string(body))
	hash := sha256.Sum256([]byte(message))

	// Sign with RSA-SHA256
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to sign: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), timestamp, nil
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

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DFNS-APPID", c.config.AppID)
	req.Header.Set("X-DFNS-NONCE", generateNonce())

	// Sign the request if we have a private key
	if c.privateKey != nil {
		signature, timestamp, err := c.signRequest(method, path, bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}
		req.Header.Set("X-DFNS-SIGNATURE", signature)
		req.Header.Set("X-DFNS-TIMESTAMP", timestamp)
	}

	resp, err := c.httpClient.Do(req)
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

// generateNonce generates a random nonce for request signing
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
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
