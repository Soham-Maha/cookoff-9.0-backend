-- name: CreateUser :one
INSERT INTO users (id, email, reg_no, password, role, round_qualified, score, name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, email, reg_no, password, role, round_qualified, score, name
FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id, email, reg_no, password, role, round_qualified, score, name
FROM users
WHERE name = $1;

-- name: GetUserById :one
SELECT id, email, reg_no, password, role, round_qualified, score, name
FROM users
WHERE id = $1;
-- name: GetAllUsers :many
SELECT id, email, reg_no, password, role, round_qualified, score, name
FROM users
-- name: UpgradeUserToRound :exec
UPDATE users
SET round_qualified = GREATEST(round_qualified, $2)
WHERE id = ANY($1::uuid[]);
-- name: BanUser :exec
UPDATE users
SET banned = true
WHERE id = ANY($1::uuid[]);
-- name: EnableRound :exec
UPDATE rounds
SET enabled = false;

UPDATE rounds
SET enabled = true
WHERE round_number = $1;
-- name: DisableRound :exec
UPDATE rounds
SET enabled = false
WHERE round_number = $1;