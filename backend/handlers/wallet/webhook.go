package wallethandlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"socialpredict/models"
	"socialpredict/services/dfns"
	"socialpredict/util"
	"time"

	"gorm.io/gorm"
)

// DFNSWebhookHandler handles incoming webhooks from DFNS
func DFNSWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook: Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	signature := r.Header.Get("X-DFNS-Signature")
	webhookSecret := os.Getenv("DFNS_WEBHOOK_SECRET")

	if webhookSecret != "" && !dfns.VerifyWebhookSignature(body, signature, webhookSecret) {
		log.Printf("Webhook: Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse webhook event
	event, err := dfns.ParseWebhookEvent(body)
	if err != nil {
		log.Printf("Webhook: Failed to parse event: %v", err)
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	log.Printf("Webhook: Received event type: %s, ID: %s", event.Kind, event.ID)

	// Handle different event types
	switch event.Kind {
	case dfns.EventTransferInbound, dfns.EventTransferConfirmed:
		handleInboundTransfer(event, body)
	case dfns.EventTransferCompleted:
		handleTransferCompleted(event)
	case dfns.EventTransferFailed:
		handleTransferFailed(event)
	default:
		log.Printf("Webhook: Unhandled event type: %s", event.Kind)
	}

	w.WriteHeader(http.StatusOK)
}

// handleInboundTransfer processes an inbound (deposit) transfer
func handleInboundTransfer(event *dfns.WebhookEvent, rawPayload []byte) {
	data, err := dfns.ParseTransferEventData(event.Data)
	if err != nil {
		log.Printf("Webhook: Failed to parse transfer event data: %v", err)
		return
	}

	// Only process inbound transfers
	if data.Direction != "Inbound" {
		log.Printf("Webhook: Skipping non-inbound transfer: %s", data.Direction)
		return
	}

	db := util.GetDB()

	// Find the wallet that received the deposit
	var wallet models.Wallet
	if err := db.Where("dfns_wallet_id = ?", data.WalletID).First(&wallet).Error; err != nil {
		log.Printf("Webhook: Wallet not found for DFNS wallet ID: %s", data.WalletID)
		return
	}

	// Check if we've already processed this transaction (idempotency)
	var existingTx models.CryptoTransaction
	if db.Where("tx_hash = ?", data.TxHash).First(&existingTx).Error == nil {
		log.Printf("Webhook: Transaction already processed: %s", data.TxHash)
		return
	}

	// Determine token symbol from contract address
	tokenSymbol := getTokenSymbolFromContract(data.Contract, wallet.ChainID, db)
	if tokenSymbol == "" {
		log.Printf("Webhook: Unknown token contract: %s on chain %d", data.Contract, wallet.ChainID)
		return
	}

	// Convert amount to credits (1:1 for stablecoins)
	decimals := dfns.GetTokenDecimals(tokenSymbol)
	amountCredits := dfns.ConvertToCredits(data.Amount, decimals)

	if amountCredits <= 0 {
		log.Printf("Webhook: Zero or negative amount after conversion: %s -> %d", data.Amount, amountCredits)
		return
	}

	// Create transaction record and credit user atomically
	now := time.Now()
	tx := models.CryptoTransaction{
		UserID:        wallet.UserID,
		WalletID:      &wallet.ID,
		Type:          models.TxTypeDeposit,
		Status:        models.TxStatusCompleted,
		ChainID:       wallet.ChainID,
		ChainName:     wallet.ChainName,
		TokenSymbol:   tokenSymbol,
		TokenAddress:  data.Contract,
		Amount:        data.Amount,
		AmountCredits: amountCredits,
		TxHash:        data.TxHash,
		FromAddress:   data.From,
		ToAddress:     data.To,
		DfnsTxID:      data.ID,
		WebhookData:   string(rawPayload),
		ProcessedAt:   &now,
	}

	// Use database transaction to atomically credit user
	dbTx := db.Begin()

	// Create transaction record
	if err := dbTx.Create(&tx).Error; err != nil {
		dbTx.Rollback()
		log.Printf("Webhook: Failed to create transaction record: %v", err)
		return
	}

	// Credit user's account balance
	var user models.User
	if err := dbTx.First(&user, wallet.UserID).Error; err != nil {
		dbTx.Rollback()
		log.Printf("Webhook: Failed to find user: %v", err)
		return
	}

	user.AccountBalance += amountCredits
	if err := dbTx.Save(&user).Error; err != nil {
		dbTx.Rollback()
		log.Printf("Webhook: Failed to credit user balance: %v", err)
		return
	}

	dbTx.Commit()
	log.Printf("Webhook: Deposit credited - User %s, Amount %d credits, TxHash %s",
		user.Username, amountCredits, data.TxHash)
}

// handleTransferCompleted processes a completed outbound transfer
func handleTransferCompleted(event *dfns.WebhookEvent) {
	data, err := dfns.ParseTransferEventData(event.Data)
	if err != nil {
		log.Printf("Webhook: Failed to parse transfer completed event: %v", err)
		return
	}

	db := util.GetDB()

	// Find the transaction by DFNS ID
	var tx models.CryptoTransaction
	if err := db.Where("dfns_tx_id = ?", data.ID).First(&tx).Error; err != nil {
		log.Printf("Webhook: Transaction not found for DFNS ID: %s", data.ID)
		return
	}

	// Update transaction status
	now := time.Now()
	tx.Status = models.TxStatusCompleted
	tx.TxHash = data.TxHash
	tx.ProcessedAt = &now

	if err := db.Save(&tx).Error; err != nil {
		log.Printf("Webhook: Failed to update transaction: %v", err)
		return
	}

	// Update associated withdrawal request
	var withdrawalReq models.WithdrawalRequest
	if tx.Type == models.TxTypeWithdrawal {
		if err := db.Where("transaction_id = ?", tx.ID).First(&withdrawalReq).Error; err == nil {
			withdrawalReq.Status = models.TxStatusCompleted
			withdrawalReq.ProcessedAt = &now
			db.Save(&withdrawalReq)
		}
	}

	log.Printf("Webhook: Transfer completed - TxID %d, TxHash %s", tx.ID, data.TxHash)
}

// handleTransferFailed processes a failed transfer
func handleTransferFailed(event *dfns.WebhookEvent) {
	data, err := dfns.ParseTransferEventData(event.Data)
	if err != nil {
		log.Printf("Webhook: Failed to parse transfer failed event: %v", err)
		return
	}

	db := util.GetDB()

	// Find the transaction by DFNS ID
	var tx models.CryptoTransaction
	if err := db.Where("dfns_tx_id = ?", data.ID).First(&tx).Error; err != nil {
		log.Printf("Webhook: Transaction not found for DFNS ID: %s", data.ID)
		return
	}

	// Update transaction status
	now := time.Now()
	tx.Status = models.TxStatusFailed
	tx.ProcessedAt = &now
	tx.ErrorMessage = "Transfer failed"

	if err := db.Save(&tx).Error; err != nil {
		log.Printf("Webhook: Failed to update transaction: %v", err)
		return
	}

	// If this was a withdrawal, refund the user
	if tx.Type == models.TxTypeWithdrawal {
		var user models.User
		if err := db.First(&user, tx.UserID).Error; err == nil {
			user.AccountBalance += tx.AmountCredits
			db.Save(&user)
			log.Printf("Webhook: Refunded %d credits to user %s due to failed withdrawal", tx.AmountCredits, user.Username)
		}

		// Update withdrawal request
		var withdrawalReq models.WithdrawalRequest
		if err := db.Where("transaction_id = ?", tx.ID).First(&withdrawalReq).Error; err == nil {
			withdrawalReq.Status = models.TxStatusFailed
			withdrawalReq.ProcessedAt = &now
			withdrawalReq.ErrorMessage = "Transfer failed on blockchain"
			db.Save(&withdrawalReq)
		}
	}

	log.Printf("Webhook: Transfer failed - TxID %d, DFNS ID %s", tx.ID, data.ID)
}

// getTokenSymbolFromContract determines the token symbol from the contract address
func getTokenSymbolFromContract(contract string, chainID int64, db *gorm.DB) string {
	var chain models.SupportedChain
	if err := db.Where("chain_id = ?", chainID).First(&chain).Error; err != nil {
		return ""
	}

	// Compare contract addresses (case-insensitive)
	if equalAddresses(contract, chain.USDCAddress) {
		return "USDC"
	}
	if equalAddresses(contract, chain.USDTAddress) {
		return "USDT"
	}

	return ""
}

// equalAddresses compares two EVM addresses case-insensitively
func equalAddresses(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
