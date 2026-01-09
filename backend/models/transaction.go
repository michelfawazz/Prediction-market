package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction type constants
const (
	TxTypeDeposit    = "DEPOSIT"
	TxTypeWithdrawal = "WITHDRAWAL"
)

// Transaction status constants
const (
	TxStatusPending   = "PENDING"
	TxStatusApproved  = "APPROVED"
	TxStatusCompleted = "COMPLETED"
	TxStatusFailed    = "FAILED"
	TxStatusRejected  = "REJECTED"
)

// CryptoTransaction tracks all deposits and withdrawals
type CryptoTransaction struct {
	gorm.Model
	ID            uint       `json:"id" gorm:"primary_key"`
	UserID        int64      `json:"userId" gorm:"index;not null"`
	WalletID      *uint      `json:"walletId" gorm:"index"`
	Type          string     `json:"type" gorm:"not null"`         // DEPOSIT or WITHDRAWAL
	Status        string     `json:"status" gorm:"index;not null"` // PENDING, APPROVED, COMPLETED, FAILED, REJECTED
	ChainID       int64      `json:"chainId"`
	ChainName     string     `json:"chainName"`
	TokenSymbol   string     `json:"tokenSymbol"`   // USDC, USDT
	TokenAddress  string     `json:"tokenAddress"`  // Contract address
	Amount        string     `json:"amount"`        // Raw amount in token decimals (string for precision)
	AmountCredits int64      `json:"amountCredits"` // Converted to platform credits (1:1 for stablecoins)
	TxHash        string     `json:"txHash" gorm:"index"`
	FromAddress   string     `json:"fromAddress"`
	ToAddress     string     `json:"toAddress"`
	DfnsTxID      string     `json:"dfnsTxId"` // DFNS transaction/request ID
	Confirmations int        `json:"confirmations" gorm:"default:0"`
	RequiredConf  int        `json:"requiredConf"`
	Fee           string     `json:"fee"`                        // Network fee
	PlatformFee   int64      `json:"platformFee" gorm:"default:0"` // Platform fee in credits
	ErrorMessage  string     `json:"errorMessage"`
	WebhookData   string     `json:"webhookData" gorm:"type:text"` // Store raw webhook data
	ProcessedAt   *time.Time `json:"processedAt"`
}

// WithdrawalRequest tracks user withdrawal requests before admin approval
type WithdrawalRequest struct {
	gorm.Model
	ID            uint       `json:"id" gorm:"primary_key"`
	UserID        int64      `json:"userId" gorm:"index;not null"`
	ChainID       int64      `json:"chainId" gorm:"not null"`
	ChainName     string     `json:"chainName" gorm:"not null"`
	TokenSymbol   string     `json:"tokenSymbol" gorm:"not null"`
	Amount        int64      `json:"amount" gorm:"not null"` // Amount in credits
	ToAddress     string     `json:"toAddress" gorm:"not null"`
	Status        string     `json:"status" gorm:"index;not null"` // PENDING, APPROVED, COMPLETED, REJECTED, FAILED
	TransactionID *uint      `json:"transactionId"`                // Link to CryptoTransaction when processed
	ErrorMessage  string     `json:"errorMessage"`
	AdminID       *int64     `json:"adminId"`                     // Admin who approved/rejected
	AdminNote     string     `json:"adminNote"`                   // Note from admin
	ProcessedAt   *time.Time `json:"processedAt"`
}

// TableName specifies the table name for CryptoTransaction
func (CryptoTransaction) TableName() string {
	return "crypto_transactions"
}

// TableName specifies the table name for WithdrawalRequest
func (WithdrawalRequest) TableName() string {
	return "withdrawal_requests"
}

// BeforeCreate hook to set creation timestamp
func (ct *CryptoTransaction) BeforeCreate(tx *gorm.DB) error {
	if ct.CreatedAt.IsZero() {
		ct.CreatedAt = time.Now()
	}
	return nil
}

// BeforeCreate hook to set creation timestamp
func (wr *WithdrawalRequest) BeforeCreate(tx *gorm.DB) error {
	if wr.CreatedAt.IsZero() {
		wr.CreatedAt = time.Now()
	}
	return nil
}

// IsDeposit returns true if the transaction is a deposit
func (ct *CryptoTransaction) IsDeposit() bool {
	return ct.Type == TxTypeDeposit
}

// IsWithdrawal returns true if the transaction is a withdrawal
func (ct *CryptoTransaction) IsWithdrawal() bool {
	return ct.Type == TxTypeWithdrawal
}

// IsPending returns true if the withdrawal request is pending
func (wr *WithdrawalRequest) IsPending() bool {
	return wr.Status == TxStatusPending
}

// CanBeApproved returns true if the withdrawal can be approved
func (wr *WithdrawalRequest) CanBeApproved() bool {
	return wr.Status == TxStatusPending
}

// CanBeRejected returns true if the withdrawal can be rejected
func (wr *WithdrawalRequest) CanBeRejected() bool {
	return wr.Status == TxStatusPending
}
