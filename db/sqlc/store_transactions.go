package db

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// CreateSplitsTxParams contains the input parameters for creating splits
type CreateSplitsTxParams struct {
	TransactionID int64
	Splits        []CreateSplitParams
}

// CreateSplitsTxResult is the result of the CreateSplitsTx operation
type CreateSplitsTxResult struct {
	Transaction Transaction
	Splits      []Split
}

// CreateSplitsTx creates multiple splits for a transaction atomically
// It validates that the splits add up to 100% of the transaction amount
func (store *SQLStore) CreateSplitsTx(ctx context.Context, arg CreateSplitsTxParams) (CreateSplitsTxResult, error) {
	var result CreateSplitsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Get the transaction to validate it exists and get the amount
		result.Transaction, err = q.GetTransactionByID(ctx, arg.TransactionID)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		// 2. Validate that splits add up to 100%
		totalPercent := decimal.NewFromInt(0)
		totalAmount := decimal.NewFromInt(0)
		for _, split := range arg.Splits {
			totalPercent = totalPercent.Add(split.SplitPercent)
			totalAmount = totalAmount.Add(split.SplitAmount)
		}

		// Check if total percent equals 1.0 (100%)
		if !totalPercent.Equal(decimal.NewFromInt(1)) {
			return fmt.Errorf("split percentages must add up to 100%%, got %s", totalPercent.String())
		}

		// Optional: Check if total amount equals transaction amount (with small tolerance for rounding)
		tolerance := decimal.NewFromFloat(0.01) // 1 cent tolerance
		amountDiff := totalAmount.Sub(result.Transaction.Amount).Abs()
		if amountDiff.GreaterThan(tolerance) {
			return fmt.Errorf("split amounts must add up to transaction amount %s, got %s (diff: %s)",
				result.Transaction.Amount.String(), totalAmount.String(), amountDiff.String())
		}

		// 3. Create all splits
		result.Splits = make([]Split, 0, len(arg.Splits))
		for _, splitParam := range arg.Splits {
			split, err := q.CreateSplit(ctx, splitParam)
			if err != nil {
				return fmt.Errorf("failed to create split: %w", err)
			}
			result.Splits = append(result.Splits, split)
		}

		return nil
	})

	return result, err
}

// UpdateTransactionSplitsTxParams contains parameters for updating all splits for a transaction
type UpdateTransactionSplitsTxParams struct {
	TransactionID int64
	Splits        []CreateSplitParams // New splits to replace existing ones
}

// UpdateTransactionSplitsTxResult is the result of the update operation
type UpdateTransactionSplitsTxResult struct {
	DeletedSplits []Split
	NewSplits     []Split
}

// UpdateTransactionSplitsTx atomically replaces all splits for a transaction
// This ensures the transaction is never left in an invalid state
func (store *SQLStore) UpdateTransactionSplitsTx(ctx context.Context, arg UpdateTransactionSplitsTxParams) (UpdateTransactionSplitsTxResult, error) {
	var result UpdateTransactionSplitsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Lock the transaction row to prevent concurrent modifications
		tx, err := q.GetTransactionByIDForUpdate(ctx, arg.TransactionID)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		// 2. Validate new splits add up to 100%
		totalPercent := decimal.NewFromInt(0)
		totalAmount := decimal.NewFromInt(0)
		for _, split := range arg.Splits {
			totalPercent = totalPercent.Add(split.SplitPercent)
			totalAmount = totalAmount.Add(split.SplitAmount)
		}

		if !totalPercent.Equal(decimal.NewFromInt(1)) {
			return fmt.Errorf("split percentages must add up to 100%%, got %s", totalPercent.String())
		}

		tolerance := decimal.NewFromFloat(0.01)
		amountDiff := totalAmount.Sub(tx.Amount).Abs()
		if amountDiff.GreaterThan(tolerance) {
			return fmt.Errorf("split amounts must add up to transaction amount %s, got %s",
				tx.Amount.String(), totalAmount.String())
		}

		// 3. Delete existing splits
		result.DeletedSplits, err = q.DeleteTransactionSplits(ctx, arg.TransactionID)
		if err != nil {
			return fmt.Errorf("failed to delete existing splits: %w", err)
		}

		// 4. Create new splits
		result.NewSplits = make([]Split, 0, len(arg.Splits))
		for _, splitParam := range arg.Splits {
			split, err := q.CreateSplit(ctx, splitParam)
			if err != nil {
				return fmt.Errorf("failed to create split: %w", err)
			}
			result.NewSplits = append(result.NewSplits, split)
		}

		return nil
	})

	return result, err
}

// DeleteTransactionWithSplitsTx deletes a transaction and its splits atomically
// (Note: CASCADE handles this automatically, but this shows explicit control)
func (store *SQLStore) DeleteTransactionWithSplitsTx(ctx context.Context, transactionID int64) error {
	return store.execTx(ctx, func(q *Queries) error {
		// Lock the transaction
		_, err := q.GetTransactionByIDForUpdate(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		// Delete splits first
		_, err = q.DeleteTransactionSplits(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("failed to delete splits: %w", err)
		}

		// Delete transaction
		_, err = q.DeleteTransaction(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("failed to delete transaction: %w", err)
		}

		return nil
	})
}
