-- name: CreateQuote :one
INSERT INTO quotes (carrier_name, service, price, deadline)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindLastQuotes :many
SELECT * FROM quotes
ORDER BY created_at DESC
LIMIT @limitQuotes::int;