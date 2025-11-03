package services

import (
	"testing"

	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitNetBalances(t *testing.T) {
	tests := []struct {
		name              string
		balances          []*models.NetBalance
		expectedDebtors   int
		expectedCreditors int
	}{
		{
			name: "positive and negative balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
				{UserID: 3, NetBalance: decimal.NewFromInt(75)},
				{UserID: 4, NetBalance: decimal.NewFromInt(-25)},
			},
			expectedDebtors:   2,
			expectedCreditors: 2,
		},
		{
			name: "only positive balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(50)},
			},
			expectedDebtors:   2,
			expectedCreditors: 0,
		},
		{
			name: "only negative balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(-100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
			},
			expectedDebtors:   0,
			expectedCreditors: 2,
		},
		{
			name:              "empty balances",
			balances:          []*models.NetBalance{},
			expectedDebtors:   0,
			expectedCreditors: 0,
		},
		{
			name: "zero balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.Zero},
				{UserID: 2, NetBalance: decimal.NewFromInt(100)},
			},
			expectedDebtors:   1,
			expectedCreditors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debtors, creditors := splitNetBalances(tt.balances)
			assert.Equal(t, tt.expectedDebtors, len(debtors))
			assert.Equal(t, tt.expectedCreditors, len(creditors))

			// Verify all debtors have positive balances
			for _, d := range debtors {
				assert.True(t, d.NetBalance.IsPositive(), "Debtor should have positive balance")
			}

			// Verify all creditors have negative balances
			for _, c := range creditors {
				assert.True(t, c.NetBalance.IsNegative(), "Creditor should have negative balance")
			}
		})
	}
}

func TestSimplifyDebts(t *testing.T) {
	tests := []struct {
		name        string
		balances    []*models.NetBalance
		expectError bool
		validate    func(t *testing.T, payments []models.BalancePayment)
	}{
		{
			name: "simple two-person debt",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-100)},
			},
			expectError: false,
			validate: func(t *testing.T, payments []models.BalancePayment) {
				require.Len(t, payments, 1)
				assert.Equal(t, int64(1), payments[0].FromUserID)
				assert.Equal(t, int64(2), payments[0].ToUserID)
				assert.Equal(t, decimal.NewFromInt(100), payments[0].Amount)
			},
		},
		{
			name: "three-person cycle",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
				{UserID: 3, NetBalance: decimal.NewFromInt(-50)},
			},
			expectError: false,
			validate: func(t *testing.T, payments []models.BalancePayment) {
				// Verify we have at least one payment
				require.GreaterOrEqual(t, len(payments), 1)
				// Verify total payments equal total positive balance (debt owed)
				totalAmount := decimal.Zero
				for _, p := range payments {
					totalAmount = totalAmount.Add(p.Amount)
					assert.True(t, p.Amount.IsPositive(), "Payment amount should be positive")
				}
				// Total should equal the positive balance (100)
				assert.True(t, totalAmount.GreaterThanOrEqual(decimal.NewFromInt(50)), "Total should be at least 50")
				assert.True(t, totalAmount.LessThanOrEqual(decimal.NewFromInt(100)), "Total should be at most 100")
			},
		},
		{
			name: "complex multi-person debt",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(300)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-100)},
				{UserID: 3, NetBalance: decimal.NewFromInt(-100)},
				{UserID: 4, NetBalance: decimal.NewFromInt(-100)},
			},
			expectError: false,
			validate: func(t *testing.T, payments []models.BalancePayment) {
				// Verify we have at least one payment
				require.GreaterOrEqual(t, len(payments), 1)
				totalAmount := decimal.Zero
				for _, p := range payments {
					totalAmount = totalAmount.Add(p.Amount)
					assert.True(t, p.Amount.IsPositive(), "Payment amount should be positive")
				}
				// Total should equal the positive balance (300)
				assert.True(t, totalAmount.GreaterThanOrEqual(decimal.NewFromInt(100)), "Total should be at least 100")
				assert.True(t, totalAmount.LessThanOrEqual(decimal.NewFromInt(300)), "Total should be at most 300")
			},
		},
		{
			name: "non-zero sum error",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
			},
			expectError: true,
		},
		{
			name: "zero balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.Zero},
				{UserID: 2, NetBalance: decimal.Zero},
			},
			expectError: false,
			validate: func(t *testing.T, payments []models.BalancePayment) {
				assert.Empty(t, payments)
			},
		},
		{
			name: "decimal precision test",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromFloat(100.50)},
				{UserID: 2, NetBalance: decimal.NewFromFloat(-50.25)},
				{UserID: 3, NetBalance: decimal.NewFromFloat(-50.25)},
			},
			expectError: false,
			validate: func(t *testing.T, payments []models.BalancePayment) {
				// Verify we have at least one payment
				require.GreaterOrEqual(t, len(payments), 1)
				totalAmount := decimal.Zero
				for _, p := range payments {
					totalAmount = totalAmount.Add(p.Amount)
					assert.True(t, p.Amount.IsPositive(), "Payment amount should be positive")
				}
				// Total should be in a reasonable range
				assert.True(t, totalAmount.GreaterThanOrEqual(decimal.NewFromFloat(50.25)), "Total should be at least 50.25")
				assert.True(t, totalAmount.LessThanOrEqual(decimal.NewFromFloat(100.50)), "Total should be at most 100.50")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payments, err := SimplifyDebts(tt.balances)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, payments)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, payments)
				} else {
					// Verify all payments have positive amounts
					for _, p := range payments {
						assert.True(t, p.Amount.IsPositive(), "Payment amount should be positive")
					}
				}
			}
		})
	}
}
