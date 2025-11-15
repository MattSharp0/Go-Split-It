/*
split queries
Table structure:
id bigserial PRIMARY KEY,
transaction_id bigint NOT NULL,
tx_amount numeric(10,2) NOT NULL,
split_percent decimal(5,4) NOT NULL,
split_amount numeric(10,2) NOT NULL,
split_user bigint,
created_at timestamptz NOT NULL DEFAULT (now()),
modified_at timestamptz NOT NULL DEFAULT (now())

*/

-- name: CreateSplit :one
INSERT INTO "splits" (transaction_id, tx_amount, split_percent, split_amount, split_user) 
VALUES ($1, (SELECT amount from transactions where id = $1), $2, $3, $4) 
RETURNING *;

-- name: GetSplitByID :one
SELECT 
    *
FROM "splits"
WHERE id = $1 
LIMIT 1;

-- name: GetSplitByIDForUpdate :one
SELECT 
    *
FROM "splits"
WHERE id = $1 
LIMIT 1
FOR UPDATE;

-- name: GetSplitsByTransactionID :many
SELECT
    *
FROM "splits"
WHERE transaction_id = $1
ORDER BY created_at desc;

-- name: GetSplitsByTransactionIDForUpdate :many
SELECT
    *
FROM "splits"
WHERE transaction_id = $1
ORDER BY created_at desc
FOR UPDATE;

-- name: GetSplitsByUser :many
SELECT * FROM "splits"
WHERE split_user = $1
ORDER BY created_at desc
LIMIT $2
OFFSET $3
FOR UPDATE;

-- name: GetSplitsByUserFiltered :many
SELECT s.* FROM "splits" s
INNER JOIN transactions t ON s.transaction_id = t.id
INNER JOIN group_members gm ON t.group_id = gm.group_id
WHERE s.split_user = $1 AND gm.user_id = $2
ORDER BY s.created_at desc
LIMIT $3
OFFSET $4;

-- name: ListSplits :many
SELECT 
    * 
FROM "splits"
ORDER BY transaction_id, created_at desc
LIMIT $1
OFFSET $2;

-- name: ListSplitsByUserGroups :many
SELECT 
    s.* 
FROM "splits" s
INNER JOIN transactions t ON s.transaction_id = t.id
INNER JOIN group_members gm ON t.group_id = gm.group_id
WHERE gm.user_id = $1
ORDER BY s.transaction_id, s.created_at desc
LIMIT $2
OFFSET $3;

-- name: ListSplitsForTransaction :many
SELECT 
    s.*
FROM "splits" s
JOIN "transactions" tx ON s.transaction_id = tx.id
WHERE s.transaction_id = $1;

-- name: UpdateSplit :one
UPDATE "splits"
SET split_percent = $2, split_amount = $3, split_user = $4
WHERE id = $1
RETURNING *;

-- name: DeleteSplit :one
DELETE FROM "splits" 
WHERE id = $1
RETURNING *; 

-- name: DeleteTransactionSplits :many
DELETE from "splits"
WHERE transaction_id = $1
RETURNING *;