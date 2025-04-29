-- name: GroupBalances :many 
SELECT gb.user_name, gb.balance
FROM group_balances gb
WHERE gb.group_id = $1
ORDER BY gb.user_name;