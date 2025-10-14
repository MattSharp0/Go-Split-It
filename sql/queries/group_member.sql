-- name: CreateGroupMember :one
INSERT INTO group_members (group_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetGroupMemberByID :one
SELECT 
  gm.*,
  g.name AS group_name,
  u.name AS user_name
FROM group_members gm
JOIN groups g ON gm.group_id = g.id 
JOIN users u ON gm.user_id = u.id
WHERE gm.id = $1 LIMIT 1;

-- name: ListGroupMembersByGroupID :many
SELECT 
  gm.*,
  g.name AS group_name,
  u.name AS user_name
FROM group_members gm
JOIN groups g ON gm.group_id = g.id
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1
ORDER BY gm.id
LIMIT $2
OFFSET $3;

-- name: UpdateGroupMember :one
UPDATE group_members
SET group_id = $1, user_id = $2
WHERE id = $3
RETURNING *;    

-- name: DeleteGroupMember :one
DELETE FROM group_members
WHERE id = $1
RETURNING *;

-- name: DeleteGroupMembersByGroupID :many
DELETE FROM group_members
WHERE group_id = $1
RETURNING *;
