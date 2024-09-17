-- :name CreateUser :one
INSERT INTO users (id, submissions, email, reg_no, password, role, round_qualified, score, name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
