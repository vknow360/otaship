-- name: GetUpdateByID :one
SELECT * FROM updates WHERE id = $1;

-- name: GetLatestActiveUpdate :one
SELECT * FROM updates 
WHERE is_active = true 
AND project_id = $1
AND platform = $2
AND runtime_version = $3
AND channel = $4
ORDER BY created_at 
DESC LIMIT 1;

-- name: ListUpdatesByProject :many
SELECT * FROM updates 
WHERE project_id = $1
ORDER BY created_at DESC 
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetUpdatesCountByProject :one
SELECT COUNT(*) FROM updates 
WHERE project_id = $1;

-- name: CreateUpdate :one
INSERT INTO updates (
    project_id,
    runtime_version,
    channel,
    rollout_percentage,
    platform,
    is_active,
    is_rollback,
    message
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateRolloutPercentage :exec
UPDATE updates 
SET rollout_percentage = $2 
WHERE id = $1;

-- name: UpdateExpoConfig :exec
UPDATE updates
SET expo_config = $1
WHERE id = $2;

-- name: DeactivateUpdates :exec
UPDATE updates 
SET is_active = false 
WHERE project_id = $1
AND channel = $2
AND runtime_version = $3
AND platform = $4
AND is_active = true;

-- name: DeleteUpdate :exec
DELETE FROM updates WHERE id = $1;

-- name: ListUpdatesPaginated :many
SELECT * FROM updates 
ORDER BY created_at DESC 
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetUpdatesCount :one
SELECT COUNT(*) FROM updates;

-- name: ActivateUpdate :exec
UPDATE updates
SET is_active = true
WHERE id = $1;

-- name: GetDownloadCountsByUpdateIDs :many
WITH stats AS (
    SELECT
        update_id,
        SUM(download_count)::bigint AS count
    FROM download_stats
    WHERE update_id = ANY($1::uuid[])
    GROUP BY update_id
),
events AS (
    SELECT
        update_id,
        COUNT(*)::bigint AS count
    FROM download_events
    WHERE update_id = ANY($1::uuid[])
    GROUP BY update_id
)
SELECT
    COALESCE(stats.update_id, events.update_id) AS update_id,
    COALESCE(stats.count, 0) + COALESCE(events.count, 0) AS count
FROM stats
FULL OUTER JOIN events
ON stats.update_id = events.update_id;
