package handlers

import (
	"fmt"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/shopspring/decimal"
)

// ValidateSplitMembersInGroup ensures all split_user IDs reference group members in the transaction's group.
func ValidateSplitMembersInGroup(splits []models.CreateSplitRequest, groupMembers []db.ListGroupMembersByGroupIDRow, groupID int64) error {

	validMemberIDs := make(map[int64]bool)
	for _, member := range groupMembers {
		validMemberIDs[member.ID] = true
	}

	for i, split := range splits {
		if split.SplitUser != nil {
			if !validMemberIDs[*split.SplitUser] {
				logger.Warn("Split is not a member of this group", "split_user", *split.SplitUser, "group_id", groupID)
				return fmt.Errorf("split[%d]: split_user %d is not a member of this group", i, *split.SplitUser)
			}
		}
	}

	return nil
}

// ValidateSplitsTotals ensures splits add up to exactly 100% and match transaction amount
func ValidateSplitsTotals(splits []models.CreateSplitRequest, transactionAmount decimal.Decimal) error {

	if len(splits) == 0 {
		return fmt.Errorf("at least one split is required")
	}

	var totalPercent decimal.Decimal
	var totalAmount decimal.Decimal

	for i, split := range splits {
		// Validate percent (should be 0.0 to 1.0)
		if split.SplitPercent.LessThan(decimal.Zero) || split.SplitPercent.GreaterThan(decimal.NewFromInt(1)) {
			logger.Debug("Split Percent not within valid range", "split_percent", split.SplitPercent)
			return fmt.Errorf("split[%d]: split_percent must be between 0.0 and 1.0", i)
		}

		// Validate amount is positive
		if split.SplitAmount.LessThanOrEqual(decimal.Zero) {
			logger.Debug("Split amount less than 0", "split_amount", split.SplitAmount)
			return fmt.Errorf("split[%d]: split_amount must be greater than 0", i)
		}

		totalPercent = totalPercent.Add(split.SplitPercent)
		totalAmount = totalAmount.Add(split.SplitAmount)
	}

	// Check percentages sum to exactly 1.0 (100%)
	expectedPercent := decimal.NewFromInt(1)
	if !totalPercent.Equal(expectedPercent) {
		logger.Debug("Split percentages do not equal 100%", "total_percent", totalPercent, "total_splits", len(splits))
		return fmt.Errorf("split percentages must sum to 1.0 (100%%), got %s", totalPercent.String())
	}

	// Check amounts sum to transaction amount (within 1 cent tolerance)
	tolerance := decimal.NewFromFloat(0.01)
	diff := totalAmount.Sub(transactionAmount).Abs()
	if diff.GreaterThan(tolerance) {
		logger.Debug("Split percentages do not equal Tx amount", "total_amount", totalAmount, "tx_amount", transactionAmount, "total_splits", len(splits))
		return fmt.Errorf("split amounts must sum to transaction amount %s, got %s (difference: %s)",
			transactionAmount.String(), totalAmount.String(), diff.String())
	}

	return nil
}
