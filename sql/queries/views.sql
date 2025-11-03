-- name: GroupBalances :many 
SELECT 
    c.user_id as creditor_id, 
    c.member_name as creditor, 
    d.user_id as debtor_id,
    d.member_name as debtor,
    gb.total_owed::numeric(10,2) as total_owed -- sum returns unconstrained numeric
FROM group_balances gb
JOIN group_members c on c.id = gb.creditor
JOIN group_members d on d.id = gb.debtor
WHERE gb.group_id = $1
ORDER BY c.member_name, d.member_name;

-- name: GroupBalancesNet :many
SELECT
    gm.user_id as user_id,
    gm.member_name as user_name,
    gbn.net_balance::numeric(10,2) as net_balance -- sum returns unconstrained numeric
FROM group_balances_net gbn
JOIN group_members gm on gm.id = gbn.user_id
WHERE gbn.group_id = $1
ORDER BY gm.member_name;

-- name: UserBalancesSummary :one
SELECT
    (SELECT 
        COALESCE(SUM(ubn.net_balance), 0)::numeric(10,2)
        FROM user_balances_net ubn 
        WHERE ubn.user_id = $1) as net_balance,
    (SELECT 
        COALESCE(SUM(CASE WHEN ubm.net_balance < 0 THEN -ubm.net_balance ELSE 0 END), 0)::numeric(10,2)
     FROM user_balances_by_member ubm 
     WHERE ubm.user_id = $1) as total_owed,
    (SELECT 
        COALESCE(SUM(CASE WHEN ubm.net_balance > 0 THEN ubm.net_balance ELSE 0 END), 0)::numeric(10,2)
     FROM user_balances_by_member ubm 
     WHERE ubm.user_id = $1) as total_owed_to_user;

-- name: UserBalancesByGroup :many
-- Returns balances by group for a specific user
-- Only includes groups where the user is a member (filtered via WHERE gm.user_id = $1)
-- This is the correct place to filter by user membership for security and performance
SELECT
    g.id as group_id,
    g.name as group_name,
    gbn.net_balance::numeric(10,2) as net_balance
FROM group_balances_net gbn
JOIN group_members gm on gm.id = gbn.user_id
JOIN groups g on g.id = gbn.group_id
WHERE gm.user_id = $1
ORDER BY g.name;

-- name: UserBalancesByMember :many
SELECT
    ubm.member_user_id as member_user_id,
    u.name as member_name,
    ubm.net_balance::numeric(10,2) as net_balance
FROM user_balances_by_member ubm
JOIN users u on u.id = ubm.member_user_id
WHERE ubm.user_id = $1
ORDER BY u.name;
