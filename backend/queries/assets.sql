-- name: GetAssetsByUpdateID :many
SELECT * FROM assets 
WHERE update_id = $1
ORDER BY file_name;

-- name: CreateAsset :exec
INSERT INTO assets (
    update_id, file_name, mime_type, key, url, hash, storage_provider, size
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteAssetByUpdateID :exec
DELETE FROM assets WHERE update_id = $1;

-- name: CloneAssets :exec
INSERT INTO assets (
    update_id, file_name, mime_type, key, url, hash, storage_provider, size
)
SELECT sqlc.arg('target_update_id'),
    a.file_name, a.mime_type, a.key, a.url, a.hash, a.storage_provider, a.size
FROM assets a
WHERE a.update_id = sqlc.arg('source_update_id');


-- name: CountOtherAssetReferences :one
SELECT COUNT(*) FROM assets
WHERE key = $1 AND update_id != $2;