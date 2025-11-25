/*
refresh token queries
Table structure:
CREATE TABLE "refresh_tokens" (
  "id" bigserial PRIMARY KEY,
  "token_hash" varchar NOT NULL,
  "user_id" bigint NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "revoked_at" timestamptz,
  "device_info" varchar,
  CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE
);
*/

-- name: CreateRefreshToken :one
INSERT INTO "refresh_tokens" (token_hash, user_id, expires_at, device_info)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM "refresh_tokens"
WHERE token_hash = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE "refresh_tokens"
SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeAllUserTokens :exec
UPDATE "refresh_tokens"
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredTokens :exec
DELETE FROM "refresh_tokens"
WHERE expires_at < now() OR (revoked_at IS NOT NULL AND revoked_at < now() - INTERVAL '30 days');

-- name: GetUserRefreshTokens :many
SELECT * FROM "refresh_tokens"
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC;

