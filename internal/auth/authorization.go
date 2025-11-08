package auth

import (
	"context"
	"errors"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
)

// IsGroupMember checks if a user is a member of a group
func IsGroupMember(ctx context.Context, store db.Store, groupID, userID int64) (bool, error) {
	// Get group members for the group
	members, err := store.ListGroupMembersByGroupID(ctx, db.ListGroupMembersByGroupIDParams{
		GroupID: groupID,
		Limit:   1000, // reasonable limit
		Offset:  0,
	})
	if err != nil {
		return false, err
	}

	// Check if user is in the members list
	for _, member := range members {
		if member.UserID != nil && *member.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

// CheckGroupMembership is a helper function to check if user is group member
// Returns error if not a member, nil if member
func CheckGroupMembership(ctx context.Context, store db.Store, groupID, userID int64) error {
	isMember, err := IsGroupMember(ctx, store, groupID, userID)
	if err != nil {
		logger.Error("Failed to check group membership", "error", err, "group_id", groupID, "user_id", userID)
		return errors.New("failed to verify group membership")
	}

	if !isMember {
		logger.Warn("User is not a member of group", "group_id", groupID, "user_id", userID)
		return errors.New("user is not a member of this group")
	}

	return nil
}

// CheckOwnUser verifies that the path user_id matches the authenticated user
func CheckOwnUser(authenticatedUserID, pathUserID int64) error {
	if authenticatedUserID != pathUserID {
		logger.Warn("User attempted to access another user's resource",
			"authenticated_user_id", authenticatedUserID,
			"path_user_id", pathUserID)
		return errors.New("forbidden: cannot access another user's resource")
	}
	return nil
}
