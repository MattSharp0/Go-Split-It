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
	logger.Debug("splitNetBalances returned", "debtors", len(debtorNetBalances), "creditors", len(creditorNetBalances))
	return debtorNetBalances, creditorNetBalances
}

func SimplifyDebts(balances []*models.NetBalance) ([]models.SimplifiedPaymentsResponse, error) {

	if sum := utils.SumNetBalances(balances); !sum.IsZero() {
		logger.Warn("Net balances do not sum to zero", "sum", sum)
		return nil, errors.New("net Balances do not sum to zero")
	}

	debtorNetBalances, creditorNetBalances := splitNetBalances(balances)

	debtorHeap := (*heaps.MaxNetBalanceHeap)(&debtorNetBalances)
	creditorHeap := (*heaps.MinNetBalanceHeap)(&creditorNetBalances)

	heap.Init(debtorHeap)
	heap.Init(creditorHeap)

	payments := make([]models.SimplifiedPaymentsResponse, debtorHeap.Len())
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
			pa = decimal.Min(creditor.NetBalance.Abs(), debtor.NetBalance)
		}
		if delta.IsPositive() {
			debtorHeap.Push(&models.NetBalance{UserID: debtor.UserID, NetBalance: delta})
			pa = decimal.Min(creditor.NetBalance.Abs(), debtor.NetBalance)
		}

		pmt := models.SimplifiedPaymentsResponse{FromUserID: debtor.UserID, ToUserID: creditor.UserID, Amount: pa}
		payments = append(payments, pmt)
		paymentCount++

	}
	logger.Debug("Simplified Payments count", "payment count", paymentCount)

	payments = payments[paymentCount:]
	return payments, nil
}
