package dfns

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Webhook event types
const (
	EventWalletCreated          = "wallet.created"
	EventWalletActivated        = "wallet.activated"
	EventTransferInbound        = "wallet.transfer.inbound"
	EventTransferOutbound       = "wallet.transfer.outbound"
	EventTransferCompleted      = "wallet.transfer.completed"
	EventTransferFailed         = "wallet.transfer.failed"
	EventTransferBroadcasted    = "wallet.transfer.broadcasted"
	EventTransferConfirmed      = "wallet.transfer.confirmed"
)

// WebhookEvent represents a webhook event from DFNS
type WebhookEvent struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
	OrgID     string          `json:"orgId"`
}

// TransferEventData represents the data for transfer webhook events
type TransferEventData struct {
	ID          string `json:"id"`
	WalletID    string `json:"walletId"`
	Network     string `json:"network"`
	Status      string `json:"status"`
	TxHash      string `json:"txHash,omitempty"`
	Direction   string `json:"direction"` // "Inbound" or "Outbound"
	Kind        string `json:"kind"`      // "Erc20", "Native"
	Symbol      string `json:"symbol,omitempty"`
	Amount      string `json:"amount"`
	From        string `json:"from"`
	To          string `json:"to"`
	Contract    string `json:"contract,omitempty"`
	Decimals    int    `json:"decimals,omitempty"`
	BlockNumber int64  `json:"blockNumber,omitempty"`
	DateCreated string `json:"dateCreated,omitempty"`
}

// WalletEventData represents the data for wallet webhook events
type WalletEventData struct {
	ID          string `json:"id"`
	Network     string `json:"network"`
	Address     string `json:"address"`
	Status      string `json:"status"`
	DateCreated string `json:"dateCreated,omitempty"`
}

// VerifyWebhookSignature validates the webhook signature using HMAC-SHA256
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// ParseWebhookEvent parses the raw webhook payload into a WebhookEvent
func ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	return &event, nil
}

// ParseTransferEventData parses the data field of a transfer webhook event
func ParseTransferEventData(data json.RawMessage) (*TransferEventData, error) {
	var transferData TransferEventData
	if err := json.Unmarshal(data, &transferData); err != nil {
		return nil, fmt.Errorf("failed to parse transfer event data: %w", err)
	}
	return &transferData, nil
}

// ParseWalletEventData parses the data field of a wallet webhook event
func ParseWalletEventData(data json.RawMessage) (*WalletEventData, error) {
	var walletData WalletEventData
	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, fmt.Errorf("failed to parse wallet event data: %w", err)
	}
	return &walletData, nil
}

// IsInboundTransfer returns true if this is an inbound transfer event
func (e *WebhookEvent) IsInboundTransfer() bool {
	return e.Kind == EventTransferInbound || e.Kind == EventTransferConfirmed
}

// IsOutboundTransfer returns true if this is an outbound transfer event
func (e *WebhookEvent) IsOutboundTransfer() bool {
	return e.Kind == EventTransferOutbound || e.Kind == EventTransferBroadcasted
}

// IsTransferComplete returns true if this is a completed transfer
func (e *WebhookEvent) IsTransferComplete() bool {
	return e.Kind == EventTransferCompleted || e.Kind == EventTransferConfirmed
}

// IsTransferFailed returns true if the transfer failed
func (e *WebhookEvent) IsTransferFailed() bool {
	return e.Kind == EventTransferFailed
}
