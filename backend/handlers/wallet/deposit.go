package wallethandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/services/dfns"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

// DepositAddressResponse represents a deposit address for a specific chain
type DepositAddressResponse struct {
	ChainID     int64  `json:"chainId"`
	ChainName   string `json:"chainName"`
	DisplayName string `json:"displayName"`
	Address     string `json:"address"`
}

// AllDepositAddressesResponse represents all deposit addresses for a user
type AllDepositAddressesResponse struct {
	Addresses []DepositAddressResponse `json:"addresses"`
}

// GetDepositAddressHandler returns the user's deposit address for a specific chain
func GetDepositAddressHandler(dfnsClient *dfns.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		vars := mux.Vars(r)
		chainName := vars["chain"]

		// Validate chain name
		if !dfns.IsValidChainName(chainName) {
			http.Error(w, "Invalid chain name", http.StatusBadRequest)
			return
		}

		// Find existing wallet for this user and chain
		var wallet models.Wallet
		result := db.Where("user_id = ? AND chain_name = ? AND is_active = ?", user.ID, chainName, true).First(&wallet)

		if result.Error != nil {
			// Wallet doesn't exist, create one via DFNS
			newWallet, err := createWalletForUser(user, chainName, dfnsClient, db)
			if err != nil {
				log.Printf("Failed to create wallet for user %s on chain %s: %v", user.Username, chainName, err)
				http.Error(w, "Failed to create deposit address", http.StatusInternalServerError)
				return
			}
			wallet = *newWallet
		}

		// Get chain info for display name
		chainInfo, ok := models.ChainInfo[chainName]
		displayName := chainName
		if ok {
			displayName = chainInfo.DisplayName
		}

		response := DepositAddressResponse{
			ChainID:     wallet.ChainID,
			ChainName:   wallet.ChainName,
			DisplayName: displayName,
			Address:     wallet.Address,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetAllDepositAddressesHandler returns deposit addresses for all supported chains
func GetAllDepositAddressesHandler(dfnsClient *dfns.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		// Get all active supported chains
		var chains []models.SupportedChain
		db.Where("is_active = ?", true).Find(&chains)

		addresses := make([]DepositAddressResponse, 0, len(chains))

		for _, chain := range chains {
			// Find or create wallet for each chain
			var wallet models.Wallet
			result := db.Where("user_id = ? AND chain_name = ? AND is_active = ?", user.ID, chain.Name, true).First(&wallet)

			if result.Error != nil {
				// Create wallet if it doesn't exist
				newWallet, err := createWalletForUser(user, chain.Name, dfnsClient, db)
				if err != nil {
					log.Printf("Failed to create wallet for user %s on chain %s: %v", user.Username, chain.Name, err)
					continue // Skip this chain but continue with others
				}
				wallet = *newWallet
			}

			addresses = append(addresses, DepositAddressResponse{
				ChainID:     wallet.ChainID,
				ChainName:   wallet.ChainName,
				DisplayName: chain.DisplayName,
				Address:     wallet.Address,
			})
		}

		response := AllDepositAddressesResponse{
			Addresses: addresses,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// createWalletForUser creates a new MPC wallet for a user on a specific chain
func createWalletForUser(user *models.User, chainName string, dfnsClient *dfns.Client, db interface{ Create(value interface{}) interface{ Error() error } }) (*models.Wallet, error) {
	// Get DFNS network name for the chain
	network := dfns.GetDFNSNetwork(chainName)
	if network == "" {
		return nil, fmt.Errorf("unknown chain: %s", chainName)
	}

	// Get chain info
	chainInfo, ok := models.ChainInfo[chainName]
	if !ok {
		return nil, fmt.Errorf("chain info not found for: %s", chainName)
	}

	// Create wallet via DFNS
	createReq := dfns.CreateWalletRequest{
		Network:    network,
		Name:       fmt.Sprintf("user-%d-%s", user.ID, chainName),
		ExternalID: fmt.Sprintf("%d", user.ID),
	}

	dfnsWallet, err := dfnsClient.CreateWallet(createReq)
	if err != nil {
		return nil, fmt.Errorf("DFNS wallet creation failed: %w", err)
	}

	// Create local wallet record
	wallet := &models.Wallet{
		UserID:       user.ID,
		DfnsWalletID: dfnsWallet.ID,
		ChainID:      chainInfo.ChainID,
		ChainName:    chainName,
		Address:      dfnsWallet.Address,
		IsActive:     true,
	}

	if err := db.Create(wallet).Error(); err != nil {
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	log.Printf("Created wallet for user %s on chain %s: %s", user.Username, chainName, wallet.Address)

	return wallet, nil
}
