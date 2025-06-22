-- name: CreateUser :one
INSERT INTO users (
    name,
    email,
    username,
    created_at,
    updated_at,
    phone_number
) VALUES (
    $1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $4
) RETURNING id, name, email, username, phone_number, created_at, updated_at;

-- name: CheckUsernameExists :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE username = $1
) AS exists;

-- name: GetUserByID :one
SELECT id, name, email, username, phone_number, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE(sqlc.narg(name), name),
    email = COALESCE(sqlc.narg(email), email),
    phone_number = COALESCE(sqlc.narg(phone_number), phone_number),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, email, username, phone_number, created_at, updated_at;

-- name: CheckEmailExistsForUpdate :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1 AND id != $2
) AS exists;