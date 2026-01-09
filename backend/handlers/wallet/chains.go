package wallethandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"
)

// ChainResponse represents a supported chain in the response
type ChainResponse struct {
	ChainID     int64  `json:"chainId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ExplorerURL string `json:"explorerUrl,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	IsActive    bool   `json:"isActive"`
}

// TokenResponse represents a supported token in the response
type TokenResponse struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	IconURL  string `json:"iconUrl,omitempty"`
	IsActive bool   `json:"isActive"`
}

// GetSupportedChainsHandler returns all supported blockchain networks
func GetSupportedChainsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	var chains []models.SupportedChain
	result := db.Where("is_active = ?", true).Order("chain_id ASC").Find(&chains)

	if result.Error != nil {
		http.Error(w, "Failed to fetch chains", http.StatusInternalServerError)
		return
	}

	response := make([]ChainResponse, len(chains))
	for i, chain := range chains {
		response[i] = ChainResponse{
			ChainID:     chain.ChainID,
			Name:        chain.Name,
			DisplayName: chain.DisplayName,
			ExplorerURL: chain.ExplorerURL,
			IconURL:     chain.IconURL,
			IsActive:    chain.IsActive,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chains": response,
	})
}

// GetSupportedTokensHandler returns all supported tokens
func GetSupportedTokensHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	var tokens []models.SupportedToken
	result := db.Where("is_active = ?", true).Order("symbol ASC").Find(&tokens)

	if result.Error != nil {
		http.Error(w, "Failed to fetch tokens", http.StatusInternalServerError)
		return
	}

	response := make([]TokenResponse, len(tokens))
	for i, token := range tokens {
		response[i] = TokenResponse{
			Symbol:   token.Symbol,
			Name:     token.Name,
			Decimals: token.Decimals,
			IconURL:  token.IconURL,
			IsActive: token.IsActive,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokens": response,
	})
}

// GetTokensForChainHandler returns the supported tokens for a specific chain
func GetTokensForChainHandler(w http.ResponseWriter, r *http.Request) {
	chainName := r.URL.Query().Get("chain")
	if chainName == "" {
		http.Error(w, "Chain name required", http.StatusBadRequest)
		return
	}

	db := util.GetDB()

	var chain models.SupportedChain
	if err := db.Where("name = ? AND is_active = ?", chainName, true).First(&chain).Error; err != nil {
		http.Error(w, "Chain not found", http.StatusNotFound)
		return
	}

	// Build list of available tokens for this chain
	tokens := []TokenResponse{}

	if chain.USDCAddress != "" {
		tokens = append(tokens, TokenResponse{
			Symbol:   "USDC",
			Name:     "USD Coin",
			Decimals: 6,
			IsActive: true,
		})
	}

	if chain.USDTAddress != "" {
		tokens = append(tokens, TokenResponse{
			Symbol:   "USDT",
			Name:     "Tether USD",
			Decimals: 6,
			IsActive: true,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chain":  chain.Name,
		"tokens": tokens,
	})
}

// GetWalletInfoHandler returns the wallet status and configuration info
func GetWalletInfoHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	// Count active chains and tokens
	var chainCount, tokenCount int64
	db.Model(&models.SupportedChain{}).Where("is_active = ?", true).Count(&chainCount)
	db.Model(&models.SupportedToken{}).Where("is_active = ?", true).Count(&tokenCount)

	response := map[string]interface{}{
		"status":          "active",
		"supportedChains": chainCount,
		"supportedTokens": tokenCount,
		"limits": map[string]int64{
			"minWithdrawal":  MinWithdrawalAmount,
			"maxWithdrawal":  MaxWithdrawalAmount,
			"dailyLimit":     DailyWithdrawalLimit,
		},
		"creditRatio": "1:1", // 1 token = 1 credit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
