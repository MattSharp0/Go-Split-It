package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionResponse struct {
	ID              int64           `json:"id"`
	GroupID         int64           `json:"group_id"`
	Name            string          `json:"name"`
	TransactionDate time.Time       `json:"transaction_date"`
	Amount          decimal.Decimal `json:"amount"`
	Category        *string         `json:"category"`
	Note            *string         `json:"note"`
	ByUser          int64           `json:"by_user"`
	CreatedAt       time.Time       `json:"created_at"`
	ModifiedAt      time.Time       `json:"modified_at"`
}

type ListTransactionResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Count        int32                 `json:"count"`
	Limit        int32                 `json:"limit"`
	Offset       int32                 `json:"offset"`
}

type CreateTransactionRequest struct {
	GroupID         int64           `json:"group_id"`
	Name            string          `json:"name"`
	TransactionDate time.Time       `json:"transaction_date"`
	Amount          decimal.Decimal `json:"amount"`
	Category        *string         `json:"category"`
	Note            *string         `json:"note"`
	ByUser          int64           `json:"by_user"`
}

type UpdateTransactionRequest struct {
	GroupID         int64           `json:"group_id"`
	Name            string          `json:"name"`
	TransactionDate time.Time       `json:"transaction_date"`
	Amount          decimal.Decimal `json:"amount"`
	Category        *string         `json:"category"`
	Note            *string         `json:"note"`
	ByUser          int64           `json:"by_user"`
}
