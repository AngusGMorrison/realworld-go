-- name: GetUserById :one
SELECT id, email, username, bio, password_hash, image_url
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, email, username, bio, password_hash, image_url
FROM users
WHERE email = ?;

-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    username,
    password_hash,
    bio,
    image_url
)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    email = COALESCE(sqlc.narg(email), email),
    username = COALESCE(sqlc.narg(username), username),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
    bio = COALESCE(sqlc.narg(bio), bio),
    image_url = COALESCE(sqlc.narg(image_url), image_url)
WHERE id = @id
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
