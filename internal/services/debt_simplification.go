package services

import (
	"container/heap"
	"errors"

	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	heaps "github.com/MattSharp0/transaction-split-go/internal/pkg/heap"
	"github.com/MattSharp0/transaction-split-go/internal/pkg/utils"
	"github.com/shopspring/decimal"
)

func splitNetBalances(balances []*models.NetBalance) ([]*models.NetBalance, []*models.NetBalance) {
	var debtorNetBalances []*models.NetBalance
	var creditorNetBalances []*models.NetBalance

	for _, balance := range balances {
		if balance.NetBalance.IsPositive() {
			debtorNetBalances = append(debtorNetBalances, balance)
		}
		if balance.NetBalance.IsNegative() {
			creditorNetBalances = append(creditorNetBalances, balance)
		}
	}
	logger.Debug("splitNetBalances returned: %s, %v")
	return debtorNetBalances, creditorNetBalances
}

func SimplifyDebts(balances []*models.NetBalance) ([]models.BalancePayment, error) {

	if sum := utils.SumNetBalances(balances); !sum.IsZero() {
		logger.Warn("Net balances do not sum to zero", "sum", sum)
		return nil, errors.New("Net Balances do not sum to zero")
	}

	creditorNetBalances, debtorNetBalances := splitNetBalances(balances)

	creditorHeap := (*heaps.MinNetBalanceHeap)(&creditorNetBalances)
	debtorHeap := (*heaps.MaxNetBalanceHeap)(&debtorNetBalances)

	heap.Init(creditorHeap)
	heap.Init(debtorHeap)

	payments := make([]models.BalancePayment, debtorHeap.Len())
	var paymentCount int

	for debtorHeap.Len() > 0 {
		debtor := heap.Pop(debtorHeap).(*models.NetBalance)
		creditor := heap.Pop(creditorHeap).(*models.NetBalance)
		delta := creditor.NetBalance.Add(debtor.NetBalance)

		var pa decimal.Decimal

		if delta.IsZero() {
			pa = debtor.NetBalance
		}
		if delta.IsNegative() {
			creditorHeap.Push(&models.NetBalance{UserID: creditor.UserID, NetBalance: delta})
			pa = delta.Abs()
		}
		if delta.IsPositive() {
			debtorHeap.Push(&models.NetBalance{UserID: debtor.UserID, NetBalance: delta})
			pa = delta
		}
		pmt := models.BalancePayment{FromUserID: debtor.UserID, ToUserID: creditor.UserID, Amount: pa}
		payments = append(payments, pmt)
		paymentCount++

	}

	payments = payments[paymentCount:]
	return payments, nil
}
