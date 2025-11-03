/*
transaction queries
Table structure:
id bigserial PRIMARY KEY,
group_id bigint NOT NULL,
name varchar NOT NULL,
transaction_date date NOT NULL DEFAULT (CURRENT_DATE),
amount numeric(10,2) NOT NULL,
category varchar,
note varchar,
by_user bigint NOT NULL,
created_at timestamptz NOT NULL DEFAULT (now()),
modified_at timestamptz NOT NULL DEFAULT (now())
*/


-- name: CreateTransaction :one
INSERT INTO "transactions" (group_id, name, transaction_date, amount, category, note, by_user)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTransactionByID :one
SELECT 
    * 
FROM "transactions"
WHERE id = $1 
LIMIT 1;

-- name: GetTransactionByIDForUpdate :one
SELECT 
    * 
FROM "transactions"
WHERE id = $1 
LIMIT 1
FOR UPDATE;

-- name: GetTransactionsByUser :many
SELECT 
    * 
FROM "transactions"
WHERE by_user = $1
ORDER BY transaction_date desc
LIMIT $2
OFFSET $3;

-- name: GetTransactionsByGroupInPeriod :many
SELECT 
    *
FROM "transactions"
WHERE 
    group_id = $1
    and transaction_date between @start_date::date and @end_date::date
ORDER BY transaction_date desc
LIMIT $2
OFFSET $3;

-- name: GetTransactionsByUserInPeriod :many
SELECT * FROM "transactions"
WHERE 
    by_user = $1 
    AND transaction_date between @start_date::date and @end_date::date
ORDER BY transaction_date desc
LIMIT $2
OFFSET $3;

-- name: ListTransactions :many
SELECT 
    * 
FROM "transactions"
ORDER BY transaction_date desc
LIMIT $1
OFFSET $2;

-- name: UpdateTransaction :one
UPDATE "transactions"
SET
    group_id = $2,
    name = $3,
    transaction_date = $4,
    amount = $5,
    category = $6,
    note = $7,
    by_user = $8
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :one
DELETE FROM "transactions" 
WHERE id = $1
RETURNING *; 