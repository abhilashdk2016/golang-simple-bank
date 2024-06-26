-- name: CreateAccount :one
INSERT INTO accounts (
    owner, balance, currency
) values (
    $1, $2, $3
) RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdates :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1 FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE owner = $1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: UpdateAccount :exec
UPDATE accounts SET balance = $3 WHERE id = $1 AND owner = $2 RETURNING *;

-- name: AddUpdateAccount :exec
UPDATE accounts SET balance = balance + sqlc.arg(amount) WHERE id = sqlc.arg(id) RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1 AND owner = $2;