-- Drop indexes before dropping views
DROP INDEX IF EXISTS splits_transaction_splituser_idx;
DROP INDEX IF EXISTS transactions_group_byuser_idx;

DROP VIEW IF EXISTS group_balances;
DROP VIEW IF EXISTS group_balances_net;
