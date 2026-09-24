CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL
);

CREATE TABLE folders (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    starred          BOOLEAN NOT NULL DEFAULT false,
    deleted          BOOLEAN NOT NULL DEFAULT false,
    deleted_at       TIMESTAMP,
    parent_folder_id TEXT REFERENCES folders(id) ON DELETE CASCADE,
    owner_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at       TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    UNIQUE (owner_id, parent_folder_id, name)
);

CREATE TABLE files (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    mime_type        TEXT NOT NULL,
    size             BIGINT NOT NULL,
    storage_key      TEXT NOT NULL,
    extension        TEXT,
    starred          BOOLEAN NOT NULL DEFAULT false,
    deleted          BOOLEAN NOT NULL DEFAULT false,
    deleted_at       TIMESTAMP,
    parent_folder_id TEXT REFERENCES folders(id) ON DELETE CASCADE,
    owner_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at       TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    -- Added via migration 00002; appended last to match ALTER TABLE column order.
    thumbnail_key    TEXT,
    UNIQUE (owner_id, parent_folder_id, name)
);

CREATE INDEX idx_files_parent_folder ON files(parent_folder_id);
CREATE INDEX idx_files_owner ON files(owner_id);
CREATE INDEX idx_files_deleted ON files(deleted);
CREATE INDEX idx_files_name ON files(name);
CREATE INDEX idx_folders_parent_folder ON folders(parent_folder_id);
CREATE INDEX idx_folders_owner ON folders(owner_id);
CREATE INDEX idx_folders_name ON folders(name);

