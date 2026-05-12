-- name: GetSetting :one
SELECT key, value FROM settings WHERE key = $1;

-- name: GetSettings :many
SELECT key, value FROM settings;

-- name: UpdateSetting :exec
INSERT INTO settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = NOW();