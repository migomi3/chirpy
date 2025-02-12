-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), Now(), Now(), $1, $2)
RETURNING *;

-- name: ClearUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserFromID :one
SELECT *
FROM users
WHERE id = $1;

-- name: UpdateLoginInfo :one
UPDATE users
SET email = $1, hashed_password = $2
WHERE id = $3
RETURNING *;