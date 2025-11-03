package mocks

import (
	"context"
	"testing"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/stretchr/testify/mock"
)

// NewMockStore creates a new instance of MockStore
func NewMockStore(t testing.TB) *MockStore {
	return &MockStore{}
}

// MockStore is a mock implementation of db.Store interface
type MockStore struct {
	mock.Mock
}

// Ensure MockStore implements the Store interface
var _ db.Store = (*MockStore)(nil)

// Querier interface methods

func (m *MockStore) CreateGroup(ctx context.Context, name string) (db.Group, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(db.Group), args.Error(1)
}

func (m *MockStore) CreateGroupMember(ctx context.Context, arg db.CreateGroupMemberParams) (db.GroupMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.GroupMember), args.Error(1)
}

func (m *MockStore) CreateSplit(ctx context.Context, arg db.CreateSplitParams) (db.Split, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Split), args.Error(1)
}

func (m *MockStore) CreateTransaction(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockStore) CreateUser(ctx context.Context, name string) (db.User, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) DeleteGroup(ctx context.Context, id int64) (db.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Group), args.Error(1)
}

func (m *MockStore) DeleteGroupMember(ctx context.Context, id int64) (db.GroupMember, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GroupMember), args.Error(1)
}

func (m *MockStore) DeleteGroupMembersByGroupID(ctx context.Context, groupID int64) ([]db.GroupMember, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.GroupMember), args.Error(1)
}

func (m *MockStore) DeleteSplit(ctx context.Context, id int64) (db.Split, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Split), args.Error(1)
}

func (m *MockStore) DeleteTransaction(ctx context.Context, id int64) (db.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockStore) DeleteTransactionSplits(ctx context.Context, transactionID int64) ([]db.Split, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) DeleteUser(ctx context.Context, id int64) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) GetGroupByID(ctx context.Context, id int64) (db.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Group), args.Error(1)
}

func (m *MockStore) GetGroupByIDForUpdate(ctx context.Context, id int64) (db.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Group), args.Error(1)
}

func (m *MockStore) GetGroupMemberByID(ctx context.Context, id int64) (db.GetGroupMemberByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GetGroupMemberByIDRow), args.Error(1)
}

func (m *MockStore) GetSplitByID(ctx context.Context, id int64) (db.Split, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Split), args.Error(1)
}

func (m *MockStore) GetSplitByIDForUpdate(ctx context.Context, id int64) (db.Split, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Split), args.Error(1)
}

func (m *MockStore) GetSplitsByTransactionID(ctx context.Context, transactionID int64) ([]db.Split, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) GetSplitsByTransactionIDForUpdate(ctx context.Context, transactionID int64) ([]db.Split, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) GetSplitsByUser(ctx context.Context, arg db.GetSplitsByUserParams) ([]db.Split, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) GetTransactionByID(ctx context.Context, id int64) (db.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockStore) GetTransactionByIDForUpdate(ctx context.Context, id int64) (db.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockStore) GetTransactionsByGroupInPeriod(ctx context.Context, arg db.GetTransactionsByGroupInPeriodParams) ([]db.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Transaction), args.Error(1)
}

func (m *MockStore) GetTransactionsByUser(ctx context.Context, arg db.GetTransactionsByUserParams) ([]db.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Transaction), args.Error(1)
}

func (m *MockStore) GetTransactionsByUserInPeriod(ctx context.Context, arg db.GetTransactionsByUserInPeriodParams) ([]db.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Transaction), args.Error(1)
}

func (m *MockStore) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) GroupBalances(ctx context.Context, groupID int64) ([]db.GroupBalancesRow, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.GroupBalancesRow), args.Error(1)
}

func (m *MockStore) GroupBalancesNet(ctx context.Context, groupID int64) ([]db.GroupBalancesNetRow, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.GroupBalancesNetRow), args.Error(1)
}

func (m *MockStore) ListGroupMembersByGroupID(ctx context.Context, arg db.ListGroupMembersByGroupIDParams) ([]db.ListGroupMembersByGroupIDRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.ListGroupMembersByGroupIDRow), args.Error(1)
}

func (m *MockStore) ListGroups(ctx context.Context, arg db.ListGroupsParams) ([]db.Group, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Group), args.Error(1)
}

func (m *MockStore) ListSplits(ctx context.Context, arg db.ListSplitsParams) ([]db.Split, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) ListSplitsForTransaction(ctx context.Context, transactionID int64) ([]db.Split, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Split), args.Error(1)
}

func (m *MockStore) ListTransactions(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Transaction), args.Error(1)
}

func (m *MockStore) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockStore) UpdateGroup(ctx context.Context, arg db.UpdateGroupParams) (db.Group, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Group), args.Error(1)
}

func (m *MockStore) UpdateGroupMember(ctx context.Context, arg db.UpdateGroupMemberParams) (db.GroupMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.GroupMember), args.Error(1)
}

func (m *MockStore) UpdateSplit(ctx context.Context, arg db.UpdateSplitParams) (db.Split, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Split), args.Error(1)
}

func (m *MockStore) UpdateTransaction(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *MockStore) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) UserBalancesByGroup(ctx context.Context, userID *int64) ([]db.UserBalancesByGroupRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.UserBalancesByGroupRow), args.Error(1)
}

func (m *MockStore) UserBalancesByMember(ctx context.Context, userID *int64) ([]db.UserBalancesByMemberRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.UserBalancesByMemberRow), args.Error(1)
}

func (m *MockStore) UserBalancesSummary(ctx context.Context, userID *int64) (db.UserBalancesSummaryRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.UserBalancesSummaryRow), args.Error(1)
}

// Transaction methods

func (m *MockStore) CreateSplitsTx(ctx context.Context, arg db.CreateSplitsTxParams) (db.CreateSplitsTxResult, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CreateSplitsTxResult), args.Error(1)
}

func (m *MockStore) UpdateTransactionSplitsTx(ctx context.Context, arg db.UpdateTransactionSplitsTxParams) (db.UpdateTransactionSplitsTxResult, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.UpdateTransactionSplitsTxResult), args.Error(1)
}

func (m *MockStore) DeleteTransactionWithSplitsTx(ctx context.Context, transactionID int64) error {
	args := m.Called(ctx, transactionID)
	return args.Error(0)
}

func (m *MockStore) CreateGroupMembersTx(ctx context.Context, arg db.CreateGroupMemberTxParams) (db.CreateGroupMemberTxResult, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CreateGroupMemberTxResult), args.Error(1)
}

func (m *MockStore) UpdateGroupMembersTx(ctx context.Context, arg db.UpdateGroupMemberTxParams) (db.UpdateGroupMemberTxResult, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.UpdateGroupMemberTxResult), args.Error(1)
}

func (m *MockStore) DeleteGroupMembersTx(ctx context.Context, groupID int64) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}
