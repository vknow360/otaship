-- name: CreateDownloadEvent :one
INSERT INTO download_events (
    update_id,
    project_id,
    device_hash,
    platform,
    channel
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRecentDownloadsByProject :many
SELECT update_id, platform, channel, COUNT(*) AS count
FROM download_events
WHERE project_id = $1
GROUP BY update_id, platform, channel
ORDER BY count DESC;

-- name: GetGlobalRecentDownloads :many
SELECT platform, channel, COUNT(*) AS count
FROM download_events
GROUP BY platform, channel
ORDER BY count DESC;

