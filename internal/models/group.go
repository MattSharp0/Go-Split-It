package models

type GroupResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListGroupResponse struct {
	Groups []GroupResponse `json:"groups"`
	Count  int32           `json:"count"`
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
}

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type UpdateGroupRequest struct {
	Name string `json:"name"`
}

// Balance response types
type BalanceResponse struct {
	Creditor  string `json:"creditor"`
	Debtor    string `json:"debtor"`
	TotalOwed int64  `json:"total_owed"`
}

type NetBalanceResponse struct {
	MemberName string `json:"member_name"`
	NetBalance int32  `json:"net_balance"`
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
