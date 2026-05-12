-- name: GetAssetByID :one
SELECT * FROM assets WHERE id = $1;

-- name: GetAssetByFileHash :one
SELECT * FROM assets WHERE file_hash = $1;

-- name: GetAssetsByUpdateID :many
SELECT * FROM assets 
WHERE update_id = $1
ORDER BY platform, file_name;

-- name: CreateAsset :exec
INSERT INTO assets (
    update_id,
    platform,
    file_name,
    mime_type,
    key,
    url,
    file_hash,
    hash,
    storage_provider
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: DeleteAssetByUpdateID :exec
DELETE FROM assets WHERE update_id = $1;

-- name: DeleteAssetByUpdateIDandPlatform :exec
DELETE FROM assets WHERE update_id = $1 AND platform = $2;

-- name: CloneAssets :exec
INSERT INTO assets (
    update_id,
    platform,
    file_name,
    mime_type,
    key,
    url,
    file_hash,
    hash,
    storage_provider
)
SELECT sqlc.arg('target_update_id'),
    a.platform,
    a.file_name,
    a.mime_type,
    a.key,
    a.url,
    a.file_hash,
    a.hash,
    a.storage_provider
FROM assets a
WHERE a.update_id = sqlc.arg('source_update_id');

-- name: CountOtherAssetReferences :one
SELECT COUNT(*) FROM assets
WHERE key = $1 AND update_id != $2;