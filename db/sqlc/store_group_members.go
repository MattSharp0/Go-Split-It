package db

import (
	"context"
	"fmt"
)

// CreateGroupMemberTxParams contains parameters for creating group members
type CreateGroupMemberTxParams struct {
	GroupID      int64
	GroupMembers []CreateGroupMemberParams
}

// CreateGroupMemberTxResult is the result of the CreateGroupMembersTx operation
type CreateGroupMemberTxResult struct {
	Group        Group
	GroupMembers []GroupMember
}

// CreateGroupMembersTx creates and adds multiple group members to a group
func (store *SQLStore) CreateGroupMembersTx(ctx context.Context, arg CreateGroupMemberTxParams) (CreateGroupMemberTxResult, error) {
	var result CreateGroupMemberTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Group, err = q.GetGroupByID(ctx, arg.GroupID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		result.GroupMembers = make([]GroupMember, 0, len(arg.GroupMembers))
		for _, groupParam := range arg.GroupMembers {
			groupMember, err := q.CreateGroupMember(ctx, groupParam)
			if err != nil {
				return fmt.Errorf("failed to create group member: %w", err)
			}
			result.GroupMembers = append(result.GroupMembers, groupMember)
		}

		return nil
	})

	return result, err
}

// UpdateGroupMemberTxParams contains parameters for updating all group members for a group
type UpdateGroupMemberTxParams struct {
	GroupID      int64
	GroupMembers []CreateGroupMemberParams // New group members to replace existing ones
}

// UpdateGroupMemberTxResult is the result of the update operation
type UpdateGroupMemberTxResult struct {
	Group          Group
	DeletedMembers []GroupMember
	NewMembers     []GroupMember
}

// UpdateGroupMembersTx atomically replaces all group members for a group
// This ensures the group is never left in an invalid state
func (store *SQLStore) UpdateGroupMembersTx(ctx context.Context, arg UpdateGroupMemberTxParams) (UpdateGroupMemberTxResult, error) {
	var result UpdateGroupMemberTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Lock the group row to prevent concurrent modifications
		result.Group, err = q.GetGroupByIDForUpdate(ctx, arg.GroupID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// 2. Delete existing group members
		result.DeletedMembers, err = q.DeleteGroupMembersByGroupID(ctx, arg.GroupID)
		if err != nil {
			return fmt.Errorf("failed to delete existing group members: %w", err)
		}

		// 3. Create new group members
		result.NewMembers = make([]GroupMember, 0, len(arg.GroupMembers))
		for _, groupParam := range arg.GroupMembers {
			groupMember, err := q.CreateGroupMember(ctx, groupParam)
			if err != nil {
				return fmt.Errorf("failed to create group member: %w", err)
			}
			result.NewMembers = append(result.NewMembers, groupMember)
		}

		return nil
	})

	return result, err
}

// DeleteGroupMembersTx deletes all group members for a group atomically
func (store *SQLStore) DeleteGroupMembersTx(ctx context.Context, groupID int64) error {
	return store.execTx(ctx, func(q *Queries) error {
		// Lock the group
		_, err := q.GetGroupByIDForUpdate(ctx, groupID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// Delete all group members
		_, err = q.DeleteGroupMembersByGroupID(ctx, groupID)
		if err != nil {
			return fmt.Errorf("failed to delete group members: %w", err)
		}

		return nil
	})
}
