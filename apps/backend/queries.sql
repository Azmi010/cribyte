-- ============================================================================
-- Auth Queries
-- ============================================================================

-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ?;

-- ============================================================================
-- Folder Queries
-- ============================================================================

-- name: CreateFolder :one
INSERT INTO folders (id, name, parent_folder_id, owner_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFolderByID :one
SELECT * FROM folders
WHERE id = ?;

-- name: GetFolderByIDAndOwner :one
SELECT * FROM folders
WHERE id = ? AND owner_id = ?;

-- name: ListRootFolders :many
SELECT * FROM folders
WHERE owner_id = ? AND parent_folder_id IS NULL AND deleted = false
ORDER BY name ASC;

-- name: ListFoldersByParent :many
SELECT * FROM folders
WHERE owner_id = ? AND parent_folder_id = ? AND deleted = false
ORDER BY name ASC;

-- name: GetExistingFolderNamesInRoot :many
SELECT name FROM folders
WHERE owner_id = ? AND parent_folder_id IS NULL AND deleted = false;

-- name: GetExistingFolderNamesInParent :many
SELECT name FROM folders
WHERE owner_id = ? AND parent_folder_id = ? AND deleted = false;

-- name: RenameFolder :one
UPDATE folders
SET name = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: MoveFolder :one
UPDATE folders
SET parent_folder_id = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: SoftDeleteFolder :one
UPDATE folders
SET deleted = true, deleted_at = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: GetFolderDescendantIDs :many
WITH RECURSIVE folder_tree AS (
    SELECT folders.id AS id FROM folders WHERE folders.id = ? AND folders.owner_id = ?
    UNION ALL
    SELECT f.id AS id FROM folders f
    JOIN folder_tree ft ON f.parent_folder_id = ft.id
)
SELECT folder_tree.id FROM folder_tree;

-- name: SoftDeleteFolderByID :exec
UPDATE folders
SET deleted = true, deleted_at = ?, updated_at = ?
WHERE id = ? AND owner_id = ?;

-- name: RestoreFolder :one
UPDATE folders
SET deleted = false, deleted_at = NULL, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: PermanentDeleteFolder :exec
DELETE FROM folders
WHERE id = ? AND owner_id = ?;

-- name: SetFolderStarred :one
UPDATE folders
SET starred = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: ListStarredFolders :many
SELECT * FROM folders
WHERE owner_id = ? AND starred = true AND deleted = false
ORDER BY name ASC;

-- name: ListTrashFolders :many
SELECT * FROM folders
WHERE owner_id = ? AND deleted = true
ORDER BY deleted_at DESC;

-- ============================================================================
-- File Queries
-- ============================================================================

-- name: CreateFile :one
INSERT INTO files (id, name, mime_type, size, storage_key, extension, parent_folder_id, owner_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files
WHERE id = ?;

-- name: GetFileByIDAndOwner :one
SELECT * FROM files
WHERE id = ? AND owner_id = ?;

-- name: ListRootFiles :many
SELECT * FROM files
WHERE owner_id = ? AND parent_folder_id IS NULL AND deleted = false
ORDER BY name ASC;

-- name: ListFilesByParent :many
SELECT * FROM files
WHERE owner_id = ? AND parent_folder_id = ? AND deleted = false
ORDER BY name ASC;

-- name: GetExistingFileNamesInRoot :many
SELECT name FROM files
WHERE owner_id = ? AND parent_folder_id IS NULL AND deleted = false;

-- name: GetExistingFileNamesInParent :many
SELECT name FROM files
WHERE owner_id = ? AND parent_folder_id = ? AND deleted = false;

-- name: RenameFile :one
UPDATE files
SET name = ?, extension = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: MoveFile :one
UPDATE files
SET parent_folder_id = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: SoftDeleteFile :one
UPDATE files
SET deleted = true, deleted_at = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: SoftDeleteFilesInFolder :exec
UPDATE files
SET deleted = true, deleted_at = ?, updated_at = ?
WHERE parent_folder_id = ? AND owner_id = ?;

-- name: RestoreFile :one
UPDATE files
SET deleted = false, deleted_at = NULL, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: PermanentDeleteFile :exec
DELETE FROM files
WHERE id = ? AND owner_id = ?;

-- name: SetFileStarred :one
UPDATE files
SET starred = ?, updated_at = ?
WHERE id = ? AND owner_id = ?
RETURNING *;

-- name: SetFileThumbnailKey :exec
UPDATE files
SET thumbnail_key = ?
WHERE id = ?;

-- name: ListStarredFiles :many
SELECT * FROM files
WHERE owner_id = ? AND starred = true AND deleted = false
ORDER BY name ASC;

-- name: ListTrashFiles :many
SELECT * FROM files
WHERE owner_id = ? AND deleted = true
ORDER BY deleted_at DESC;

-- name: SearchFilesByName :many
SELECT * FROM files
WHERE owner_id = ? AND deleted = false AND name LIKE ?
ORDER BY name ASC;

-- name: GetStorageKeysByFolderID :many
SELECT storage_key FROM files
WHERE parent_folder_id = ? AND owner_id = ?;

-- name: GetTotalStorageUsedByOwner :one
SELECT COALESCE(SUM(size), 0) AS total_size
FROM files
WHERE owner_id = ? AND deleted = false;
