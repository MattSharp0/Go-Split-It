/*
Group queries
Table structure:
id bigserial PRIMARY KEY,
name varchar NOT NULL
*/

-- name: CreateGroup :one
INSERT INTO "groups" (name)
VALUES ($1)
RETURNING *;

-- name: GetGroupByID :one
SELECT 
  *
FROM "groups"
WHERE id = $1 LIMIT 1;  

-- name: GetGroupByIDForUpdate :one
SELECT 
  *
FROM "groups"
WHERE id = $1 
LIMIT 1
FOR UPDATE;

-- name: ListGroups :many
SELECT 
  *
FROM "groups"
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateGroup :one
UPDATE "groups"
SET name = $1
WHERE id = $2
RETURNING *;    

-- name: DeleteGroup :one
DELETE FROM "groups"
WHERE id = $1
RETURNING *;