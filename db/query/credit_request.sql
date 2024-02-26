-- name: GetCreditRequestsByUsername :many
SELECT * FROM credit_requests
WHERE username = $1;

-- name: CreateCreditRequest :one
INSERT INTO credit_requests (
  username,
  reason,
  amount,
  currency
)
VALUES (
  $1, $2, $3, $4
) RETURNING *;


-- name: GetUsersPendingCreditRequests :many
SELECT * FROM credit_requests
WHERE status = 'pending';

-- name: GetPendingCreditRequestById :one
SELECT * FROM credit_requests
WHERE id = $1 and status = 'pending';

-- name: CancelCreditRequestById :one
UPDATE credit_requests
SET status = 'cancelled'
WHERE id = $1
RETURNING *;
