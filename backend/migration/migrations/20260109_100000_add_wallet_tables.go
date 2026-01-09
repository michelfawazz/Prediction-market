package migrations

import (
	"log"

	"socialpredict/logger"
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

func init() {
	err := migration.Register("20260109100000", func(db *gorm.DB) error {
		// Migrate the Wallet model
		if err := db.AutoMigrate(&models.Wallet{}); err != nil {
			return err
		}

		// Migrate the SupportedChain model
		if err := db.AutoMigrate(&models.SupportedChain{}); err != nil {
			return err
		}

		// Migrate the SupportedToken model
		if err := db.AutoMigrate(&models.SupportedToken{}); err != nil {
			return err
		}

		// Migrate the CryptoTransaction model
		if err := db.AutoMigrate(&models.CryptoTransaction{}); err != nil {
			return err
		}

		// Migrate the WithdrawalRequest model
		if err := db.AutoMigrate(&models.WithdrawalRequest{}); err != nil {
			return err
		}

		// Seed default supported chains
		chains := []models.SupportedChain{
			{
				ChainID:          1,
				Name:             "ethereum",
				DisplayName:      "Ethereum",
				USDCAddress:      "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
				USDTAddress:      "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				MinConfirmations: 12,
				IsActive:         true,
			},
			{
				ChainID:          137,
				Name:             "polygon",
				DisplayName:      "Polygon",
				USDCAddress:      "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359", // Native USDC on Polygon
				USDTAddress:      "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
				MinConfirmations: 128,
				IsActive:         true,
			},
			{
				ChainID:          8453,
				Name:             "base",
				DisplayName:      "Base",
				USDCAddress:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // Native USDC on Base
				USDTAddress:      "",                                           // No native USDT on Base
				MinConfirmations: 12,
				IsActive:         true,
			},
			{
				ChainID:          42161,
				Name:             "arbitrum",
				DisplayName:      "Arbitrum",
				USDCAddress:      "0xaf88d065e77c8cC2239327C5EDb3A432268e5831", // Native USDC on Arbitrum
				USDTAddress:      "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9",
				MinConfirmations: 12,
				IsActive:         true,
			},
		}

		for _, chain := range chains {
			// Use FirstOrCreate to avoid duplicates
			if err := db.Where("chain_id = ?", chain.ChainID).FirstOrCreate(&chain).Error; err != nil {
				return err
			}
		}

		// Seed default supported tokens
		tokens := []models.SupportedToken{
			{
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
				IsActive: true,
			},
			{
				Symbol:   "USDT",
				Name:     "Tether USD",
				Decimals: 6,
				IsActive: true,
			},
		}

		for _, token := range tokens {
			// Use FirstOrCreate to avoid duplicates
			if err := db.Where("symbol = ?", token.Symbol).FirstOrCreate(&token).Error; err != nil {
				return err
			}
		}

		return nil
	})

	// In init() functions, registration failure is a critical startup error
	if err != nil {
		logger.LogError("migrations", "init", err)
		log.Fatalf("Failed to register migration 20260109100000: %v", err)
	}
}
