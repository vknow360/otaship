-- name: CreateDownloadEvent :one
INSERT INTO download_events (
    update_id,
    project_id,
    device_hash,
    platform,
    channel
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTotalDownloadsByProject :one
SELECT COUNT(*) AS count
FROM download_events
WHERE project_id = $1;

-- name: GetRecentDownloadsByProject :one
SELECT COUNT(*) AS count
FROM download_events
WHERE project_id = sqlc.arg('project_id')
  AND timestamp > sqlc.arg('since');

-- name: GetDownloadsByPlatform :many
SELECT platform, COUNT(*) AS count
FROM download_events
WHERE project_id = $1
GROUP BY platform
ORDER BY count DESC;

-- name: GetDownloadsByChannel :many
SELECT channel, COUNT(*) AS count
FROM download_events
WHERE project_id = $1
GROUP BY channel
ORDER BY count DESC;

-- name: GetDownloadsByUpdate :many
SELECT 
    update_id,
    COUNT(*) AS download_count,
    COUNT(DISTINCT device_hash) AS unique_devices
FROM download_events
WHERE project_id = $1
GROUP BY update_id
ORDER BY MAX(timestamp) DESC;

-- name: GetGlobalTotalDownloads :one
SELECT COUNT(*) AS count FROM download_events;

-- name: GetGlobalRecentDownloads :one
SELECT COUNT(*) AS count 
FROM download_events 
WHERE timestamp > $1;

-- name: GetGlobalDownloadsByPlatform :many
SELECT platform, COUNT(*) AS count 
FROM download_events 
GROUP BY platform 
ORDER BY count DESC;

-- name: GetGlobalDownloadsByChannel :many
SELECT channel, COUNT(*) AS count 
FROM download_events 
GROUP BY channel 
ORDER BY count DESC;

-- name: PruneOldDownloadEvents :exec
DELETE FROM download_events WHERE timestamp < $1;