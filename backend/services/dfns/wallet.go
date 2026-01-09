package dfns

import (
	"encoding/json"
	"fmt"
)

// CreateWalletRequest represents a request to create a new wallet
type CreateWalletRequest struct {
	Network    string `json:"network"`              // "EthereumMainnet", "Polygon", "Base", "ArbitrumOne"
	Name       string `json:"name,omitempty"`       // Optional wallet name
	ExternalID string `json:"externalId,omitempty"` // Our internal reference (e.g., user ID)
}

// WalletResponse represents a wallet from the DFNS API
type WalletResponse struct {
	ID          string `json:"id"`
	Network     string `json:"network"`
	Address     string `json:"address"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status"`
	DateCreated string `json:"dateCreated"`
	ExternalID  string `json:"externalId,omitempty"`
}

// WalletListResponse represents a list of wallets
type WalletListResponse struct {
	Items      []WalletResponse `json:"items"`
	NextCursor string           `json:"nextCursor,omitempty"`
}

// CreateWallet creates a new MPC wallet on a specific network
func (c *Client) CreateWallet(req CreateWalletRequest) (*WalletResponse, error) {
	path := "/wallets"

	respBody, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	var wallet WalletResponse
	if err := json.Unmarshal(respBody, &wallet); err != nil {
		return nil, fmt.Errorf("failed to parse wallet response: %w", err)
	}

	return &wallet, nil
}

// GetWallet retrieves a wallet by its ID
func (c *Client) GetWallet(walletID string) (*WalletResponse, error) {
	path := fmt.Sprintf("/wallets/%s", walletID)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	var wallet WalletResponse
	if err := json.Unmarshal(respBody, &wallet); err != nil {
		return nil, fmt.Errorf("failed to parse wallet response: %w", err)
	}

	return &wallet, nil
}

// ListWallets lists all wallets, optionally filtered by network
func (c *Client) ListWallets(network string) (*WalletListResponse, error) {
	path := "/wallets"
	if network != "" {
		path = fmt.Sprintf("/wallets?network=%s", network)
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	var list WalletListResponse
	if err := json.Unmarshal(respBody, &list); err != nil {
		return nil, fmt.Errorf("failed to parse wallet list response: %w", err)
	}

	return &list, nil
}

// GetWalletBalance retrieves the balance of a specific asset in a wallet
func (c *Client) GetWalletBalance(walletID string) (*WalletBalanceResponse, error) {
	path := fmt.Sprintf("/wallets/%s/assets", walletID)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	var balance WalletBalanceResponse
	if err := json.Unmarshal(respBody, &balance); err != nil {
		return nil, fmt.Errorf("failed to parse balance response: %w", err)
	}

	return &balance, nil
}

// WalletBalanceResponse represents the assets in a wallet
type WalletBalanceResponse struct {
	Items []WalletAsset `json:"items"`
}

// WalletAsset represents an asset held in a wallet
type WalletAsset struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Balance  string `json:"balance"`
	Decimals int    `json:"decimals"`
	Contract string `json:"contract,omitempty"` // For ERC20 tokens
}
