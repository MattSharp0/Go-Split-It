package handlers

import (
	"testing"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test group members
func createTestGroupMembers(ids []int64, groupID int64) []db.ListGroupMembersByGroupIDRow {
	members := make([]db.ListGroupMembersByGroupIDRow, len(ids))
	for i, id := range ids {
		memberName := "Member " + string(rune('A'+int(i)))
		userID := int64Ptr(id)
		members[i] = db.ListGroupMembersByGroupIDRow{
			ID:         id,
			GroupID:    groupID,
			MemberName: &memberName,
			UserID:     userID,
		}
	}
	return members
}

func TestValidateSplitMembersInGroup(t *testing.T) {
	groupID := int64(1)
	validMemberIDs := []int64{1, 2, 3}
	groupMembers := createTestGroupMembers(validMemberIDs, groupID)

	tests := []struct {
		name        string
		splits      []models.CreateSplitRequest
		members     []db.ListGroupMembersByGroupIDRow
		groupID     int64
		expectError bool
		errorMsg    string
	}{
		{
			name: "all splits have valid group members",
			splits: []models.CreateSplitRequest{
				{SplitUser: int64Ptr(1), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitUser: int64Ptr(2), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: false,
		},
		{
			name: "splits with nil split_user are valid",
			splits: []models.CreateSplitRequest{
				{SplitUser: nil, SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitUser: int64Ptr(1), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: false,
		},
		{
			name: "all splits have nil split_user",
			splits: []models.CreateSplitRequest{
				{SplitUser: nil, SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitUser: nil, SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: false,
		},
		{
			name: "split_user not in group members",
			splits: []models.CreateSplitRequest{
				{SplitUser: int64Ptr(1), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitUser: int64Ptr(999), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: true,
			errorMsg:    "split_user 999 is not a member of this group",
		},
		{
			name: "first split has invalid member",
			splits: []models.CreateSplitRequest{
				{SplitUser: int64Ptr(999), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitUser: int64Ptr(2), SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: true,
			errorMsg:    "split[0]: split_user 999 is not a member of this group",
		},
		{
			name:        "empty splits array",
			splits:      []models.CreateSplitRequest{},
			members:     groupMembers,
			groupID:     groupID,
			expectError: false,
		},
		{
			name: "empty group members",
			splits: []models.CreateSplitRequest{
				{SplitUser: int64Ptr(1), SplitPercent: decimal.NewFromFloat(1.0), SplitAmount: decimal.NewFromInt(100)},
			},
			members:     []db.ListGroupMembersByGroupIDRow{},
			groupID:     groupID,
			expectError: true,
			errorMsg:    "split_user 1 is not a member of this group",
		},
		{
			name: "multiple invalid members",
			splits: []models.CreateSplitRequest{
				{SplitUser: int64Ptr(1), SplitPercent: decimal.NewFromFloat(0.33), SplitAmount: decimal.NewFromInt(33)},
				{SplitUser: int64Ptr(999), SplitPercent: decimal.NewFromFloat(0.33), SplitAmount: decimal.NewFromInt(33)},
				{SplitUser: int64Ptr(888), SplitPercent: decimal.NewFromFloat(0.34), SplitAmount: decimal.NewFromInt(34)},
			},
			members:     groupMembers,
			groupID:     groupID,
			expectError: true,
			errorMsg:    "split[1]: split_user 999 is not a member of this group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSplitMembersInGroup(tt.splits, tt.members, tt.groupID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSplitsTotals(t *testing.T) {
	tests := []struct {
		name              string
		splits            []models.CreateSplitRequest
		transactionAmount decimal.Decimal
		expectError       bool
		errorMsg          string
	}{
		{
			name: "valid splits - 50/50 split",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       false,
		},
		{
			name: "invalid splits - 33/33/33 split",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
			},
			transactionAmount: decimal.NewFromFloat(100.00),
			expectError:       true,
		},
		{
			name: "valid splits - 33/33/34 split",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333334), SplitAmount: decimal.NewFromFloat(33.34)},
			},
			transactionAmount: decimal.NewFromFloat(100.00),
			expectError:       false,
		},
		{
			name: "valid splits - single split",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromInt(1), SplitAmount: decimal.NewFromInt(100)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       false,
		},
		{
			name: "valid splits - with cents tolerance",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33.33)},
				{SplitPercent: decimal.NewFromFloat(0.333334), SplitAmount: decimal.NewFromFloat(33.35)}, // 0.01 over
			},
			transactionAmount: decimal.NewFromFloat(100.00),
			expectError:       false, // Within 1 cent tolerance
		},
		{
			name:              "empty splits array",
			splits:            []models.CreateSplitRequest{},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "at least one split is required",
		},
		{
			name: "percentages don't sum to 100%",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.4), SplitAmount: decimal.NewFromInt(50)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split percentages must sum to 1.0 (100%)",
		},
		{
			name: "percentages sum to more than 100%",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.6), SplitAmount: decimal.NewFromInt(50)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split percentages must sum to 1.0 (100%)",
		},
		{
			name: "amounts don't sum to transaction amount",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(40)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split amounts must sum to transaction amount",
		},
		{
			name: "amounts exceed transaction amount beyond tolerance",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromFloat(50.02)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromFloat(50.02)},
			},
			transactionAmount: decimal.NewFromFloat(100.00),
			expectError:       true,
			errorMsg:          "split amounts must sum to transaction amount",
		},
		{
			name: "negative split percent",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(-0.1), SplitAmount: decimal.NewFromInt(50)},
				{SplitPercent: decimal.NewFromFloat(1.1), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split_percent must be between 0.0 and 1.0",
		},
		{
			name: "split percent greater than 1.0",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(1.5), SplitAmount: decimal.NewFromInt(100)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split_percent must be between 0.0 and 1.0",
		},
		{
			name: "zero split amount",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.Zero},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split_amount must be greater than 0",
		},
		{
			name: "negative split amount",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(-10)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split_amount must be greater than 0",
		},
		{
			name: "valid splits - large transaction amount",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33333.33)},
				{SplitPercent: decimal.NewFromFloat(0.333333), SplitAmount: decimal.NewFromFloat(33333.33)},
				{SplitPercent: decimal.NewFromFloat(0.333334), SplitAmount: decimal.NewFromFloat(33333.34)},
			},
			transactionAmount: decimal.NewFromFloat(100000.00),
			expectError:       false,
		},
		{
			name: "valid splits - small transaction amount",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromFloat(0.50)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromFloat(0.50)},
			},
			transactionAmount: decimal.NewFromFloat(1.00),
			expectError:       false,
		},
		{
			name: "valid splits - uneven split",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.25), SplitAmount: decimal.NewFromFloat(25.00)},
				{SplitPercent: decimal.NewFromFloat(0.75), SplitAmount: decimal.NewFromFloat(75.00)},
			},
			transactionAmount: decimal.NewFromFloat(100.00),
			expectError:       false,
		},
		{
			name: "percentages sum to exactly 1.0 but amounts don't match",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(60)},
				{SplitPercent: decimal.NewFromFloat(0.5), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split amounts must sum to transaction amount",
		},
		{
			name: "amounts sum correctly but percentages don't",
			splits: []models.CreateSplitRequest{
				{SplitPercent: decimal.NewFromFloat(0.3), SplitAmount: decimal.NewFromInt(50)},
				{SplitPercent: decimal.NewFromFloat(0.6), SplitAmount: decimal.NewFromInt(50)},
			},
			transactionAmount: decimal.NewFromInt(100),
			expectError:       true,
			errorMsg:          "split percentages must sum to 1.0 (100%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSplitsTotals(tt.splits, tt.transactionAmount)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
