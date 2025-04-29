/*
user queries
Table strcuture:
CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);
*/

-- name: CreateUser :one
INSERT INTO "users" (name) 
VALUES ($1) 
RETURNING *;

-- name: GetUserByID :one
SELECT 
  * 
FROM "users"
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT 
  * 
FROM "users"
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE "users"
SET name = $1
WHERE id = $2
RETURNING *;

-- name: DeleteUser :one
DELETE FROM "users"
WHERE id = $1
RETURNING *; 