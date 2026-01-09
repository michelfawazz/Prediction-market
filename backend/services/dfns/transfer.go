package dfns

import (
	"encoding/json"
	"fmt"
)

// TransferKind represents the type of transfer
const (
	TransferKindErc20  = "Erc20"
	TransferKindNative = "Native"
)

// TransferRequest represents a request to transfer assets from a wallet
type TransferRequest struct {
	Kind     string `json:"kind"`              // "Erc20" for token transfers, "Native" for ETH
	To       string `json:"to"`                // Destination address
	Contract string `json:"contract,omitempty"` // Token contract address (for Erc20)
	Amount   string `json:"amount"`            // Amount in smallest unit (wei/base units)
}

// TransferResponse represents a transfer initiated via DFNS
type TransferResponse struct {
	ID          string `json:"id"`
	WalletID    string `json:"walletId"`
	Network     string `json:"network"`
	Status      string `json:"status"` // "Pending", "Executing", "Confirmed", "Failed"
	TxHash      string `json:"txHash,omitempty"`
	DateCreated string `json:"dateCreated"`
}

// TransferListResponse represents a list of transfers
type TransferListResponse struct {
	Items      []TransferResponse `json:"items"`
	NextCursor string             `json:"nextCursor,omitempty"`
}

// InitiateTransfer starts a transfer from a wallet
func (c *Client) InitiateTransfer(walletID string, req TransferRequest) (*TransferResponse, error) {
	path := fmt.Sprintf("/wallets/%s/transfers", walletID)

	respBody, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate transfer: %w", err)
	}

	var transfer TransferResponse
	if err := json.Unmarshal(respBody, &transfer); err != nil {
		return nil, fmt.Errorf("failed to parse transfer response: %w", err)
	}

	return &transfer, nil
}

// GetTransfer retrieves a transfer by its ID
func (c *Client) GetTransfer(walletID, transferID string) (*TransferResponse, error) {
	path := fmt.Sprintf("/wallets/%s/transfers/%s", walletID, transferID)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	var transfer TransferResponse
	if err := json.Unmarshal(respBody, &transfer); err != nil {
		return nil, fmt.Errorf("failed to parse transfer response: %w", err)
	}

	return &transfer, nil
}

// ListTransfers lists all transfers for a wallet
func (c *Client) ListTransfers(walletID string) (*TransferListResponse, error) {
	path := fmt.Sprintf("/wallets/%s/transfers", walletID)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}

	var list TransferListResponse
	if err := json.Unmarshal(respBody, &list); err != nil {
		return nil, fmt.Errorf("failed to parse transfer list response: %w", err)
	}

	return &list, nil
}

// BroadcastTransactionRequest represents a request to broadcast a signed transaction
type BroadcastTransactionRequest struct {
	Transaction string `json:"transaction"` // Signed transaction data
}

// BroadcastTransaction broadcasts a pre-signed transaction
func (c *Client) BroadcastTransaction(walletID string, req BroadcastTransactionRequest) (*TransferResponse, error) {
	path := fmt.Sprintf("/wallets/%s/transactions", walletID)

	respBody, err := c.doRequest("POST", path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	var transfer TransferResponse
	if err := json.Unmarshal(respBody, &transfer); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return &transfer, nil
}
