-- name: CreateAPIKey :one
INSERT INTO api_keys (project_id, name, key_hash, key_suffix)
VALUES (sqlc.arg('project_id'), sqlc.arg('name'), sqlc.arg('key_hash'), sqlc.arg('key_suffix'))
RETURNING *;

-- name: GetAPIKeyBySuffix :one
SELECT id, project_id, key_hash 
FROM api_keys 
WHERE key_suffix = $1;

-- name: ListAPIKeys :many
SELECT id, name, key_suffix, created_at, last_used_at 
FROM api_keys 
WHERE project_id = $1 
ORDER BY created_at DESC;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys 
WHERE id = sqlc.arg('id') AND project_id = sqlc.arg('project_id');

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys 
SET last_used_at = CURRENT_TIMESTAMP 
WHERE id = $1;