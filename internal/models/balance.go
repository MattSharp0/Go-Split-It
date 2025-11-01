package models

import "github.com/shopspring/decimal"

// Debt simplification balance types
type NetBalance struct {
	UserID     int64           `json:"user_id"`
	NetBalance decimal.Decimal `json:"net_balance"`
}

type BalancePayment struct {
	FromUserID int64           `json:"from_user_id"`
	ToUserID   int64           `json:"to_user_id"`
	Amount     decimal.Decimal `json:"amount"`
}

// Balance response types
type BalanceResponse struct {
	CreditorID int64           `json:"creditor_id"`
	Creditor   string          `json:"creditor"`
	DebtorID   int64           `json:"debtor_id"`
	Debtor     string          `json:"debtor"`
	TotalOwed  decimal.Decimal `json:"total_owed"`
}

type NetBalanceResponse struct {
	UserID     int64           `json:"user_id"`
	MemberName string          `json:"member_name"`
	NetBalance decimal.Decimal `json:"net_balance"`
}

type GroupBalancesResponse struct {
	GroupID         int64                `json:"group_id"`
	Balances        []BalanceResponse    `json:"balances"`
	NetBalances     []NetBalanceResponse `json:"net_balances"`
	SimplifiedOwes  []BalanceResponse    `json:"simplified_owes"`
	Count           int32                `json:"count"`
	NetCount        int32                `json:"net_count"`
	SimplifiedCount int32                `json:"simplified_count"`
}
