-- name: AggregateDownloadEvents :exec
INSERT INTO download_stats (project_id, update_id, platform, channel, date, download_count)
SELECT 
    project_id, update_id, platform, channel,
    timestamp::date AS date,
    COUNT(*) AS download_count
FROM download_events
WHERE timestamp < $1
GROUP BY project_id, update_id, platform, channel, timestamp::date
ON CONFLICT (project_id, update_id, platform, channel, date)
DO UPDATE SET
    download_count = download_stats.download_count + EXCLUDED.download_count;


-- name: DeleteAggregatedEvents :exec
DELETE FROM download_events WHERE timestamp < $1;

-- Project-level stats 
-- name: GetTotalDownloadsByProject :many
SELECT update_id, platform, channel, SUM(download_count)::bigint as count
FROM download_stats
WHERE project_id = $1
GROUP BY update_id, platform, channel
ORDER BY count DESC;

-- name: GetTotalDownloadsByUpdateID :one
SELECT (
    COALESCE((
        SELECT SUM(download_count)
        FROM download_stats ds
        WHERE ds.update_id = $1
    ), 0)
    +
    (
        SELECT COUNT(*)
        FROM download_events de
        WHERE de.update_id = $1
    )
)::bigint AS count;

-- Global stats
-- name: GetTotalDownloadStats :many
SELECT platform, channel, SUM(download_count)::bigint as count
FROM download_stats
GROUP BY platform, channel
ORDER BY count DESC;
