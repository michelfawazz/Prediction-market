package wallethandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/services/dfns"
	"socialpredict/util"
	"time"
)

// WithdrawalLimits defines the withdrawal limits
const (
	MinWithdrawalAmount   = 10    // Minimum credits per withdrawal
	MaxWithdrawalAmount   = 10000 // Maximum credits per single withdrawal
	DailyWithdrawalLimit  = 50000 // Maximum credits per day
)

// WithdrawalRequestBody represents the request body for initiating a withdrawal
type WithdrawalRequestBody struct {
	ChainName   string `json:"chainName"`
	TokenSymbol string `json:"tokenSymbol"`
	Amount      int64  `json:"amount"`    // Amount in credits
	ToAddress   string `json:"toAddress"` // External wallet address
}

// WithdrawalResponse represents the response for a withdrawal request
type WithdrawalResponse struct {
	RequestID   uint      `json:"requestId"`
	Status      string    `json:"status"`
	ChainName   string    `json:"chainName"`
	TokenSymbol string    `json:"tokenSymbol"`
	Amount      int64     `json:"amount"`
	ToAddress   string    `json:"toAddress"`
	CreatedAt   time.Time `json:"createdAt"`
	Message     string    `json:"message,omitempty"`
}

// InitiateWithdrawalHandler processes a withdrawal request
func InitiateWithdrawalHandler(dfnsClient *dfns.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := util.GetDB()
		user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
		if httperr != nil {
			http.Error(w, httperr.Error(), httperr.StatusCode)
			return
		}

		var req WithdrawalRequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate chain name
		if !dfns.IsValidChainName(req.ChainName) {
			http.Error(w, "Invalid chain name", http.StatusBadRequest)
			return
		}

		// Validate token symbol
		if !dfns.IsValidTokenSymbol(req.TokenSymbol) {
			http.Error(w, "Invalid token symbol. Supported: USDC, USDT", http.StatusBadRequest)
			return
		}

		// Validate destination address format
		if !dfns.IsValidEVMAddress(req.ToAddress) {
			http.Error(w, "Invalid destination address", http.StatusBadRequest)
			return
		}

		// Validate minimum withdrawal
		if req.Amount < MinWithdrawalAmount {
			http.Error(w, "Minimum withdrawal is 10 credits", http.StatusBadRequest)
			return
		}

		// Validate maximum single withdrawal
		if req.Amount > MaxWithdrawalAmount {
			http.Error(w, "Maximum single withdrawal is 10,000 credits", http.StatusBadRequest)
			return
		}

		// Check user has sufficient balance
		if user.AccountBalance < req.Amount {
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
			return
		}

		// Check daily withdrawal limit
		if err := checkDailyWithdrawalLimit(db, user.ID, req.Amount); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get chain info
		chainInfo, ok := models.ChainInfo[req.ChainName]
		if !ok {
			http.Error(w, "Chain configuration not found", http.StatusInternalServerError)
			return
		}

		// Use transaction to debit balance and create request atomically
		tx := db.Begin()

		// Debit user balance immediately
		user.AccountBalance -= req.Amount
		if err := tx.Save(user).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to process withdrawal", http.StatusInternalServerError)
			return
		}

		// Create withdrawal request in PENDING state (awaiting admin approval)
		withdrawalReq := models.WithdrawalRequest{
			UserID:      user.ID,
			ChainID:     chainInfo.ChainID,
			ChainName:   req.ChainName,
			TokenSymbol: req.TokenSymbol,
			Amount:      req.Amount,
			ToAddress:   req.ToAddress,
			Status:      models.TxStatusPending,
		}

		if err := tx.Create(&withdrawalReq).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Failed to create withdrawal request", http.StatusInternalServerError)
			return
		}

		tx.Commit()

		response := WithdrawalResponse{
			RequestID:   withdrawalReq.ID,
			Status:      withdrawalReq.Status,
			ChainName:   req.ChainName,
			TokenSymbol: req.TokenSymbol,
			Amount:      req.Amount,
			ToAddress:   req.ToAddress,
			CreatedAt:   withdrawalReq.CreatedAt,
			Message:     "Withdrawal request submitted. It will be processed after admin approval.",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserWithdrawalsHandler returns the user's withdrawal requests
func GetUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var requests []models.WithdrawalRequest
	db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(50).Find(&requests)

	type WithdrawalListItem struct {
		ID          uint       `json:"id"`
		ChainName   string     `json:"chainName"`
		TokenSymbol string     `json:"tokenSymbol"`
		Amount      int64      `json:"amount"`
		ToAddress   string     `json:"toAddress"`
		Status      string     `json:"status"`
		CreatedAt   time.Time  `json:"createdAt"`
		ProcessedAt *time.Time `json:"processedAt,omitempty"`
	}

	items := make([]WithdrawalListItem, len(requests))
	for i, req := range requests {
		items[i] = WithdrawalListItem{
			ID:          req.ID,
			ChainName:   req.ChainName,
			TokenSymbol: req.TokenSymbol,
			Amount:      req.Amount,
			ToAddress:   req.ToAddress,
			Status:      req.Status,
			CreatedAt:   req.CreatedAt,
			ProcessedAt: req.ProcessedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"withdrawals": items,
	})
}

// checkDailyWithdrawalLimit checks if the user has exceeded daily withdrawal limits
func checkDailyWithdrawalLimit(db interface{ Model(value interface{}) interface{ Where(query interface{}, args ...interface{}) interface{ Select(query interface{}, args ...interface{}) interface{ Scan(dest interface{}) error } } } }, userID int64, amount int64) error {
	today := time.Now().Truncate(24 * time.Hour)

	var dailyTotal int64
	db.Model(&models.WithdrawalRequest{}).
		Where("user_id = ? AND created_at >= ? AND status != ?", userID, today, models.TxStatusRejected).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&dailyTotal)

	if dailyTotal+amount > DailyWithdrawalLimit {
		return &WithdrawalLimitError{
			Message:    "Daily withdrawal limit exceeded",
			DailyLimit: DailyWithdrawalLimit,
			Used:       dailyTotal,
			Requested:  amount,
		}
	}

	return nil
}

// WithdrawalLimitError represents a withdrawal limit error
type WithdrawalLimitError struct {
	Message    string
	DailyLimit int64
	Used       int64
	Requested  int64
}

func (e *WithdrawalLimitError) Error() string {
	return e.Message
}
