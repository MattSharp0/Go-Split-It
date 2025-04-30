-- name: GroupBalances :many 
SELECT 
    c.member_name as creditor, 
    d.member_name as debtor,
    gb.total_owed
FROM group_balances gb
JOIN group_members c on c.id = gb.creditor
JOIN group_members d on d.id = gb.debtor
WHERE gb.group_id = $1
ORDER BY c.member_name, d.member_name;

-- name: GroupBalancesSimplified :many 
SELECT 
    c.member_name as creditor,
    d.member_name as debtor, 
    gbs.total_owed
FROM group_balances_simple gbs
JOIN group_members c on c.id = gbs.creditor
JOIN group_members d on d.id = gbs.debtor
WHERE gbs.group_id = $1
ORDER BY c.member_name, d.member_name;

-- name: GroupBalancesNet :many
SELECT
    gm.member_name,
    gbn.net_balance
FROM group_balances_net gbn
JOIN group_members gm on gm.id = gbn.user
WHERE gbn.group_id = $1
ORDER BY gm.member_name;
