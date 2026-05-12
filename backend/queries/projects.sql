-- name: GetProjectBySlug :one
SELECT id, slug, name, description, created_at FROM projects WHERE slug = $1;

-- name: GetProjectByID :one
SELECT id, slug, name, description, created_at FROM projects WHERE id = $1;

-- name: ListProjects :many
SELECT id, slug, name, description, created_at FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (slug, name, description) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: DeleteProject :one
DELETE FROM projects WHERE id = $1 RETURNING *;

-- name: UpdateProject :one
UPDATE projects 
SET name = $2, description = $3 
WHERE id = $1 
RETURNING *;