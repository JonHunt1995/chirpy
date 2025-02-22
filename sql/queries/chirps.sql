-- up.sql
CREATE TABLE chirps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllChirps :many
SELECT *
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT *
FROM chirps
WHERE ID = $1;

-- down.sql
DROP TABLE chirps;