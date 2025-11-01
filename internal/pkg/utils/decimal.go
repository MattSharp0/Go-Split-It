package utils

import (
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/shopspring/decimal"
)

// Sum the balance values of the given NetBalances.
// Used to validate net balance structs prior to use in debt simplification
func SumNetBalances(balances ...[]*models.NetBalance) decimal.Decimal {
	var sum decimal.Decimal
	for _, v := range balances {
		for _, b := range v {
			sum = sum.Add(b.NetBalance)
		}
	}
	logger.Debug("SumNetBalance returned %d", sum.IntPart())
	return sum

}
