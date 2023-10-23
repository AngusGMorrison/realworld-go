-- name: GetUserById :one
SELECT id, email, username, bio, password_hash, image_url, updated_at
FROM users
WHERE id = $1;

-- name: UserExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE id = $1
);

-- name: GetUserByEmail :one
SELECT id, email, username, bio, password_hash, image_url, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    username,
    password_hash,
    bio,
    image_url
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    email = COALESCE(sqlc.narg(email), email),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
    bio = COALESCE(sqlc.narg(bio), bio),
    image_url = COALESCE(sqlc.narg(image_url), image_url)
WHERE id = $1
    AND updated_at = $2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
