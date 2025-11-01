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
