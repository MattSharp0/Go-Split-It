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

type SimplifiedPaymentsResponse struct {
	FromUserID int64           `json:"from_user_id"`
	ToUserID   int64           `json:"to_user_id"`
	Amount     decimal.Decimal `json:"amount"`
}

type GroupBalancesResponse struct {
	GroupID                 int64                        `json:"group_id"`
	Balances                []BalanceResponse            `json:"balances"`
	NetBalances             []NetBalanceResponse         `json:"net_balances"`
	SimplifiedPayments      []SimplifiedPaymentsResponse `json:"simplified_payments"`
	Count                   int32                        `json:"count"`
	NetCount                int32                        `json:"net_count"`
	SimplifiedPaymentsCount int32                        `json:"simplified_payments_count"`
}

// User balance response types
type UserBalancesSummaryResponse struct {
	TotalOwed       decimal.Decimal `json:"total_owed"`
	TotalOwedToUser decimal.Decimal `json:"total_owed_to_user"`
	NetBalance      decimal.Decimal `json:"net_balance"`
}

type UserBalanceByGroupResponse struct {
	GroupID         int64           `json:"group_id"`
	GroupName       string          `json:"group_name"`
	NetBalance      decimal.Decimal `json:"net_balance"`
	TotalOwed       decimal.Decimal `json:"total_owed"`         // negative net_balance amounts
	TotalOwedToUser decimal.Decimal `json:"total_owed_to_user"` // positive net_balance amounts
}

type UserBalanceByMemberResponse struct {
	MemberID        int64           `json:"member_id"`
	MemberName      string          `json:"member_name"`
	NetBalance      decimal.Decimal `json:"net_balance"`
	TotalOwed       decimal.Decimal `json:"total_owed"`         // negative net_balance
	TotalOwedToUser decimal.Decimal `json:"total_owed_to_user"` // positive net_balance
}

type UserBalancesResponse struct {
	UserID           int64                         `json:"user_id"`
	Summary          UserBalancesSummaryResponse   `json:"summary"`
	BalancesByGroup  []UserBalanceByGroupResponse  `json:"balances_by_group"`
	BalancesByMember []UserBalanceByMemberResponse `json:"balances_by_member"`
	GroupCount       int32                         `json:"group_count"`
	MemberCount      int32                         `json:"member_count"`
}
