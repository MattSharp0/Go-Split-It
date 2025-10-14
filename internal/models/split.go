package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type SplitResponse struct {
	ID            int64           `json:"id"`
	TransactionID int64           `json:"transaction_id"`
	TxAmount      decimal.Decimal `json:"tx_amount"`
	SplitPercent  decimal.Decimal `json:"split_percent"`
	SplitAmount   decimal.Decimal `json:"split_amount"`
	SplitUser     *int64          `json:"split_user"`
	CreatedAt     time.Time       `json:"created_at"`
	ModifiedAt    time.Time       `json:"modified_at"`
}

type ListSplitResponse struct {
	Splits []SplitResponse `json:"splits"`
	Count  int32           `json:"count"`
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
}

type CreateSplitRequest struct {
	TransactionID int64           `json:"transaction_id"`
	SplitPercent  decimal.Decimal `json:"split_percent"`
	SplitAmount   decimal.Decimal `json:"split_amount"`
	SplitUser     *int64          `json:"split_user"`
}

type UpdateSplitRequest struct {
	SplitPercent decimal.Decimal `json:"split_percent"`
	SplitAmount  decimal.Decimal `json:"split_amount"`
	SplitUser    *int64          `json:"split_user"`
}
