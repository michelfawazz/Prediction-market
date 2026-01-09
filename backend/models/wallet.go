package models

import (
	"time"

	"gorm.io/gorm"
)

// Wallet represents a user's MPC wallet on a specific chain
type Wallet struct {
	gorm.Model
	ID           uint   `json:"id" gorm:"primary_key"`
	UserID       int64  `json:"userId" gorm:"index;not null"`
	DfnsWalletID string `json:"dfnsWalletId" gorm:"unique;not null"` // DFNS wallet identifier
	ChainID      int64  `json:"chainId" gorm:"not null"`             // EVM chain ID (1=ETH, 137=Polygon, 8453=Base, 42161=Arbitrum)
	ChainName    string `json:"chainName" gorm:"not null"`           // Human readable: "ethereum", "polygon", "base", "arbitrum"
	Address      string `json:"address" gorm:"index;not null"`       // Wallet address on this chain
	IsActive     bool   `json:"isActive" gorm:"default:true"`
}

// SupportedChain represents a blockchain that the platform supports
type SupportedChain struct {
	gorm.Model
	ID               uint   `json:"id" gorm:"primary_key"`
	ChainID          int64  `json:"chainId" gorm:"unique;not null"`
	Name             string `json:"name" gorm:"not null"`        // "ethereum", "polygon", "base", "arbitrum"
	DisplayName      string `json:"displayName" gorm:"not null"` // "Ethereum Mainnet"
	RpcURL           string `json:"rpcUrl"`
	ExplorerURL      string `json:"explorerUrl"`
	USDCAddress      string `json:"usdcAddress"`                   // USDC contract address on this chain
	USDTAddress      string `json:"usdtAddress"`                   // USDT contract address on this chain
	MinConfirmations int    `json:"minConfirmations" gorm:"default:12"`
	IsActive         bool   `json:"isActive" gorm:"default:true"`
	IconURL          string `json:"iconUrl"`
}

// SupportedToken represents a token that can be deposited/withdrawn
type SupportedToken struct {
	gorm.Model
	ID       uint   `json:"id" gorm:"primary_key"`
	Symbol   string `json:"symbol" gorm:"not null"` // "USDC", "USDT"
	Name     string `json:"name" gorm:"not null"`   // "USD Coin", "Tether USD"
	Decimals int    `json:"decimals" gorm:"default:6"`
	IsActive bool   `json:"isActive" gorm:"default:true"`
	IconURL  string `json:"iconUrl"`
}

// ChainInfo maps chain names to their IDs and DFNS network names
var ChainInfo = map[string]struct {
	ChainID     int64
	DfnsNetwork string
	DisplayName string
}{
	"ethereum": {ChainID: 1, DfnsNetwork: "EthereumMainnet", DisplayName: "Ethereum"},
	"polygon":  {ChainID: 137, DfnsNetwork: "Polygon", DisplayName: "Polygon"},
	"base":     {ChainID: 8453, DfnsNetwork: "Base", DisplayName: "Base"},
	"arbitrum": {ChainID: 42161, DfnsNetwork: "ArbitrumOne", DisplayName: "Arbitrum"},
}

// TokenInfo maps token symbols to their decimals
var TokenInfo = map[string]struct {
	Name     string
	Decimals int
}{
	"USDC": {Name: "USD Coin", Decimals: 6},
	"USDT": {Name: "Tether USD", Decimals: 6},
}

// GetChainID returns the chain ID for a given chain name
func GetChainID(chainName string) int64 {
	if info, ok := ChainInfo[chainName]; ok {
		return info.ChainID
	}
	return 0
}

// GetDfnsNetwork returns the DFNS network name for a given chain name
func GetDfnsNetwork(chainName string) string {
	if info, ok := ChainInfo[chainName]; ok {
		return info.DfnsNetwork
	}
	return ""
}

// GetChainNameByID returns the chain name for a given chain ID
func GetChainNameByID(chainID int64) string {
	for name, info := range ChainInfo {
		if info.ChainID == chainID {
			return name
		}
	}
	return ""
}

// TableName specifies the table name for Wallet
func (Wallet) TableName() string {
	return "wallets"
}

// TableName specifies the table name for SupportedChain
func (SupportedChain) TableName() string {
	return "supported_chains"
}

// TableName specifies the table name for SupportedToken
func (SupportedToken) TableName() string {
	return "supported_tokens"
}

// BeforeCreate hook to set creation timestamp
func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	return nil
}
