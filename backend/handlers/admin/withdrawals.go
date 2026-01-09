package adminhandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/services/dfns"
	"socialpredict/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// WithdrawalRequestItem represents a withdrawal request in the admin list
type WithdrawalRequestItem struct {
	ID          uint       `json:"id"`
	UserID      int64      `json:"userId"`
	Username    string     `json:"username"`
	ChainName   string     `json:"chainName"`
	TokenSymbol string     `json:"tokenSymbol"`
	Amount      int64      `json:"amount"`
	ToAddress   string     `json:"toAddress"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	AdminNote   string     `json:"adminNote,omitempty"`
}

// ListWithdrawalRequestsHandler returns all withdrawal requests for admin review
func ListWithdrawalRequestsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	// Validate admin token
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query params
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 50
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit

	// Build query
	query := db.Model(&models.WithdrawalRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var requests []models.WithdrawalRequest
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&requests)

	// Build response with user info
	items := make([]WithdrawalRequestItem, len(requests))
	for i, req := range requests {
		// Get username
		var user models.User
		db.Select("username").First(&user, req.UserID)

		items[i] = WithdrawalRequestItem{
			ID:          req.ID,
			UserID:      req.UserID,
			Username:    user.Username,
			ChainName:   req.ChainName,
			TokenSymbol: req.TokenSymbol,
			Amount:      req.Amount,
			ToAddress:   req.ToAddress,
			Status:      req.Status,
			CreatedAt:   req.CreatedAt,
			ProcessedAt: req.ProcessedAt,
			AdminNote:   req.AdminNote,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"withdrawals": items,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// ApproveWithdrawalRequest represents the request body for approving a withdrawal
type ApproveWithdrawalRequest struct {
	Note string `json:"note,omitempty"` // Optional admin note
}

// ApproveWithdrawalHandler approves a withdrawal request and initiates the DFNS transfer
func ApproveWithdrawalHandler(dfnsClient *dfns.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()

		// Validate admin token and get admin user
		admin, err := middleware.ValidateTokenAndGetUser(r, db)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if admin.UserType != "ADMIN" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Get withdrawal ID from URL
		vars := mux.Vars(r)
		withdrawalIDStr := vars["id"]
		withdrawalID, parseErr := strconv.ParseUint(withdrawalIDStr, 10, 32)
		if parseErr != nil {
			http.Error(w, "Invalid withdrawal ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var req ApproveWithdrawalRequest
		json.NewDecoder(r.Body).Decode(&req) // Optional, ignore errors

		// Find the withdrawal request
		var withdrawalReq models.WithdrawalRequest
		if dbErr := db.First(&withdrawalReq, withdrawalID).Error; dbErr != nil {
			http.Error(w, "Withdrawal request not found", http.StatusNotFound)
			return
		}

		// Check if can be approved
		if !withdrawalReq.CanBeApproved() {
			http.Error(w, fmt.Sprintf("Cannot approve withdrawal in status: %s", withdrawalReq.Status), http.StatusBadRequest)
			return
		}

		// Find the user's wallet for this chain
		var wallet models.Wallet
		if err := db.Where("user_id = ? AND chain_id = ? AND is_active = ?",
			withdrawalReq.UserID, withdrawalReq.ChainID, true).First(&wallet).Error; err != nil {
			http.Error(w, "User wallet not found for this chain", http.StatusBadRequest)
			return
		}

		// Get chain info for token contract address
		var chain models.SupportedChain
		if err := db.Where("chain_id = ?", withdrawalReq.ChainID).First(&chain).Error; err != nil {
			http.Error(w, "Chain configuration not found", http.StatusInternalServerError)
			return
		}

		// Determine token contract address
		var tokenContract string
		switch withdrawalReq.TokenSymbol {
		case "USDC":
			tokenContract = chain.USDCAddress
		case "USDT":
			tokenContract = chain.USDTAddress
		default:
			http.Error(w, "Unsupported token", http.StatusBadRequest)
			return
		}

		if tokenContract == "" {
			http.Error(w, "Token not available on this chain", http.StatusBadRequest)
			return
		}

		// Convert credits to token amount
		decimals := dfns.GetTokenDecimals(withdrawalReq.TokenSymbol)
		tokenAmount := dfns.CreditsToTokenAmount(withdrawalReq.Amount, decimals)

		// Initiate transfer via DFNS
		transferReq := dfns.TransferRequest{
			Kind:     dfns.TransferKindErc20,
			To:       withdrawalReq.ToAddress,
			Contract: tokenContract,
			Amount:   tokenAmount,
		}

		dfnsTransfer, transferErr := dfnsClient.InitiateTransfer(wallet.DfnsWalletID, transferReq)
		if transferErr != nil {
			log.Printf("Admin: Failed to initiate DFNS transfer for withdrawal %d: %v", withdrawalReq.ID, transferErr)
			http.Error(w, "Failed to initiate blockchain transfer", http.StatusInternalServerError)
			return
		}

		// Create crypto transaction record
		now := time.Now()
		cryptoTx := models.CryptoTransaction{
			UserID:        withdrawalReq.UserID,
			WalletID:      &wallet.ID,
			Type:          models.TxTypeWithdrawal,
			Status:        models.TxStatusApproved,
			ChainID:       withdrawalReq.ChainID,
			ChainName:     withdrawalReq.ChainName,
			TokenSymbol:   withdrawalReq.TokenSymbol,
			TokenAddress:  tokenContract,
			Amount:        tokenAmount,
			AmountCredits: withdrawalReq.Amount,
			ToAddress:     withdrawalReq.ToAddress,
			DfnsTxID:      dfnsTransfer.ID,
		}

		db.Create(&cryptoTx)

		// Update withdrawal request
		withdrawalReq.Status = models.TxStatusApproved
		withdrawalReq.TransactionID = &cryptoTx.ID
		withdrawalReq.AdminID = &admin.ID
		withdrawalReq.AdminNote = req.Note
		withdrawalReq.ProcessedAt = &now

		db.Save(&withdrawalReq)

		log.Printf("Admin: Approved withdrawal %d by admin %s, DFNS transfer ID: %s",
			withdrawalReq.ID, admin.Username, dfnsTransfer.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Withdrawal approved and transfer initiated",
			"withdrawalId":  withdrawalReq.ID,
			"transactionId": cryptoTx.ID,
			"dfnsTransferId": dfnsTransfer.ID,
			"status":        withdrawalReq.Status,
		})
	}
}

// RejectWithdrawalRequest represents the request body for rejecting a withdrawal
type RejectWithdrawalRequest struct {
	Reason string `json:"reason"` // Required reason for rejection
}

// RejectWithdrawalHandler rejects a withdrawal request and refunds the user
func RejectWithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	// Validate admin token and get admin user
	admin, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if admin.UserType != "ADMIN" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// Get withdrawal ID from URL
	vars := mux.Vars(r)
	withdrawalIDStr := vars["id"]
	withdrawalID, parseErr := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if parseErr != nil {
		http.Error(w, "Invalid withdrawal ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req RejectWithdrawalRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Rejection reason is required", http.StatusBadRequest)
		return
	}

	// Find the withdrawal request
	var withdrawalReq models.WithdrawalRequest
	if err := db.First(&withdrawalReq, withdrawalID).Error; err != nil {
		http.Error(w, "Withdrawal request not found", http.StatusNotFound)
		return
	}

	// Check if can be rejected
	if !withdrawalReq.CanBeRejected() {
		http.Error(w, fmt.Sprintf("Cannot reject withdrawal in status: %s", withdrawalReq.Status), http.StatusBadRequest)
		return
	}

	// Start transaction for atomic refund
	tx := db.Begin()

	// Refund the user's balance
	var user models.User
	if err := tx.First(&user, withdrawalReq.UserID).Error; err != nil {
		tx.Rollback()
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	user.AccountBalance += withdrawalReq.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to refund user balance", http.StatusInternalServerError)
		return
	}

	// Update withdrawal request
	now := time.Now()
	withdrawalReq.Status = models.TxStatusRejected
	withdrawalReq.AdminID = &admin.ID
	withdrawalReq.AdminNote = req.Reason
	withdrawalReq.ErrorMessage = req.Reason
	withdrawalReq.ProcessedAt = &now

	if err := tx.Save(&withdrawalReq).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update withdrawal request", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	log.Printf("Admin: Rejected withdrawal %d by admin %s, reason: %s, refunded %d credits to user %s",
		withdrawalReq.ID, admin.Username, req.Reason, withdrawalReq.Amount, user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Withdrawal rejected and credits refunded",
		"withdrawalId":   withdrawalReq.ID,
		"refundedAmount": withdrawalReq.Amount,
		"status":         withdrawalReq.Status,
	})
}

// GetWithdrawalDetailsHandler returns details for a specific withdrawal request
func GetWithdrawalDetailsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	// Validate admin token
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get withdrawal ID from URL
	vars := mux.Vars(r)
	withdrawalIDStr := vars["id"]
	withdrawalID, err := strconv.ParseUint(withdrawalIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid withdrawal ID", http.StatusBadRequest)
		return
	}

	// Find the withdrawal request
	var withdrawalReq models.WithdrawalRequest
	if err := db.First(&withdrawalReq, withdrawalID).Error; err != nil {
		http.Error(w, "Withdrawal request not found", http.StatusNotFound)
		return
	}

	// Get user info
	var user models.User
	db.Select("username, account_balance").First(&user, withdrawalReq.UserID)

	// Get associated transaction if exists
	var cryptoTx *models.CryptoTransaction
	if withdrawalReq.TransactionID != nil {
		var tx models.CryptoTransaction
		if err := db.First(&tx, *withdrawalReq.TransactionID).Error; err == nil {
			cryptoTx = &tx
		}
	}

	response := map[string]interface{}{
		"withdrawal": map[string]interface{}{
			"id":          withdrawalReq.ID,
			"userId":      withdrawalReq.UserID,
			"username":    user.Username,
			"chainName":   withdrawalReq.ChainName,
			"tokenSymbol": withdrawalReq.TokenSymbol,
			"amount":      withdrawalReq.Amount,
			"toAddress":   withdrawalReq.ToAddress,
			"status":      withdrawalReq.Status,
			"createdAt":   withdrawalReq.CreatedAt,
			"processedAt": withdrawalReq.ProcessedAt,
			"adminNote":   withdrawalReq.AdminNote,
			"error":       withdrawalReq.ErrorMessage,
		},
		"user": map[string]interface{}{
			"currentBalance": user.AccountBalance,
		},
	}

	if cryptoTx != nil {
		response["transaction"] = map[string]interface{}{
			"id":       cryptoTx.ID,
			"txHash":   cryptoTx.TxHash,
			"dfnsTxId": cryptoTx.DfnsTxID,
			"status":   cryptoTx.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWithdrawalStatsHandler returns withdrawal statistics for admin dashboard
func GetWithdrawalStatsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()

	// Validate admin token
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var pendingCount, approvedCount, completedCount, rejectedCount, failedCount int64
	var totalPendingAmount, totalCompletedAmount int64

	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusPending).Count(&pendingCount)
	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusApproved).Count(&approvedCount)
	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusCompleted).Count(&completedCount)
	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusRejected).Count(&rejectedCount)
	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusFailed).Count(&failedCount)

	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusPending).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalPendingAmount)
	db.Model(&models.WithdrawalRequest{}).Where("status = ?", models.TxStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalCompletedAmount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending": map[string]interface{}{
			"count":  pendingCount,
			"amount": totalPendingAmount,
		},
		"approved": map[string]interface{}{
			"count": approvedCount,
		},
		"completed": map[string]interface{}{
			"count":  completedCount,
			"amount": totalCompletedAmount,
		},
		"rejected": map[string]interface{}{
			"count": rejectedCount,
		},
		"failed": map[string]interface{}{
			"count": failedCount,
		},
	})
}
