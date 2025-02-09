-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, Now(), Now(), $2, $3, NULL)
RETURNING *;

-- name: ClearTokens :exec
DELETE FROM refresh_tokens;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = Now(), updated_at = Now()
WHERE token = $1;