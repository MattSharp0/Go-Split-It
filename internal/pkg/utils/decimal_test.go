package utils

import (
	"testing"

	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSumNetBalances(t *testing.T) {
	tests := []struct {
		name     string
		balances []*models.NetBalance
		expected decimal.Decimal
	}{
		{
			name: "positive balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(50)},
			},
			expected: decimal.NewFromInt(150),
		},
		{
			name: "negative balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(-100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
			},
			expected: decimal.NewFromInt(-150),
		},
		{
			name: "mixed balances",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-50)},
				{UserID: 3, NetBalance: decimal.NewFromInt(25)},
			},
			expected: decimal.NewFromInt(75),
		},
		{
			name:     "empty slice",
			balances: []*models.NetBalance{},
			expected: decimal.Zero,
		},
		{
			name: "single value",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(42)},
			},
			expected: decimal.NewFromInt(42),
		},
		{
			name: "zero sum",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(100)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-100)},
			},
			expected: decimal.Zero,
		},
		{
			name: "decimal precision",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromFloat(100.50)},
				{UserID: 2, NetBalance: decimal.NewFromFloat(-50.25)},
			},
			expected: decimal.NewFromFloat(50.25),
		},
		{
			name: "large numbers",
			balances: []*models.NetBalance{
				{UserID: 1, NetBalance: decimal.NewFromInt(1000000)},
				{UserID: 2, NetBalance: decimal.NewFromInt(-500000)},
			},
			expected: decimal.NewFromInt(500000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SumNetBalances(tt.balances)
			assert.True(t, result.Equal(tt.expected), "Expected %s, got %s", tt.expected.String(), result.String())
		})
	}
}
