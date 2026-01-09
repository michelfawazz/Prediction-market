package wallethandlers

import (
	"encoding/json"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"
)

// TransactionListResponse represents a paginated list of transactions
type TransactionListResponse struct {
	Transactions []TransactionItem `json:"transactions"`
	Total        int64             `json:"total"`
	Page         int               `json:"page"`
	PageSize     int               `json:"pageSize"`
}

// TransactionItem represents a single transaction in the list
type TransactionItem struct {
	ID            uint       `json:"id"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	ChainName     string     `json:"chainName"`
	TokenSymbol   string     `json:"tokenSymbol"`
	Amount        int64      `json:"amount"`
	TxHash        string     `json:"txHash,omitempty"`
	FromAddress   string     `json:"fromAddress,omitempty"`
	ToAddress     string     `json:"toAddress,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
}

// GetTransactionHistoryHandler returns the user's crypto transaction history
func GetTransactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// Optional filters
	txType := r.URL.Query().Get("type")   // DEPOSIT, WITHDRAWAL
	status := r.URL.Query().Get("status") // PENDING, COMPLETED, etc.

	// Build query
	query := db.Model(&models.CryptoTransaction{}).Where("user_id = ?", user.ID)

	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var transactions []models.CryptoTransaction
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions)

	// Map to response items
	items := make([]TransactionItem, len(transactions))
	for i, tx := range transactions {
		items[i] = TransactionItem{
			ID:          tx.ID,
			Type:        tx.Type,
			Status:      tx.Status,
			ChainName:   tx.ChainName,
			TokenSymbol: tx.TokenSymbol,
			Amount:      tx.AmountCredits,
			TxHash:      tx.TxHash,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			CreatedAt:   tx.CreatedAt,
			ProcessedAt: tx.ProcessedAt,
		}
	}

	response := TransactionListResponse{
		Transactions: items,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTransactionByIDHandler returns a specific transaction by ID
func GetTransactionByIDHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	// Get transaction ID from URL
	txIDStr := r.URL.Query().Get("id")
	if txIDStr == "" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	txID, err := strconv.ParseUint(txIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var tx models.CryptoTransaction
	if err := db.Where("id = ? AND user_id = ?", txID, user.ID).First(&tx).Error; err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	response := TransactionItem{
		ID:          tx.ID,
		Type:        tx.Type,
		Status:      tx.Status,
		ChainName:   tx.ChainName,
		TokenSymbol: tx.TokenSymbol,
		Amount:      tx.AmountCredits,
		TxHash:      tx.TxHash,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		CreatedAt:   tx.CreatedAt,
		ProcessedAt: tx.ProcessedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
