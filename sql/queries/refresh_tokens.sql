-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $1, updated_at = $2
WHERE token = $3;