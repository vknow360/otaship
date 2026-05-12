-- 1. Settings (global configuration)
CREATE TABLE settings (
    key VARCHAR(30) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. API Keys (scoped to projects)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_suffix VARCHAR(16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

-- 4. Updates
CREATE TABLE updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    runtime_version TEXT NOT NULL,
    channel TEXT NOT NULL,
    rollout_percentage INT NOT NULL DEFAULT 100 CHECK (rollout_percentage BETWEEN 0 AND 100),
    platform TEXT NOT NULL CHECK (platform IN ('android', 'ios')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_rollback BOOLEAN NOT NULL DEFAULT false,
    message TEXT,
    expo_config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. Assets
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    update_id UUID NOT NULL REFERENCES updates(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    key TEXT NOT NULL,
    url TEXT NOT NULL,
    file_hash TEXT NOT NULL,
    hash TEXT NOT NULL,
    storage_provider TEXT NOT NULL DEFAULT 'cloudinary'
);

-- 6. Download events (for analytics)
CREATE TABLE download_events (
    id BIGSERIAL PRIMARY KEY,
    update_id UUID NOT NULL REFERENCES updates(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
    device_hash TEXT NOT NULL,
    platform TEXT NOT NULL,
    channel TEXT NOT NULL
);

-- 7. Constraints
ALTER TABLE updates ADD CONSTRAINT updates_channel_check CHECK (channel ~ '^[a-z0-9][a-z0-9_-]{0,32}$');

-- 8. Indexes for performance
CREATE INDEX idx_updates_manifest ON updates(project_id, channel, runtime_version, is_active, created_at DESC);
CREATE INDEX idx_assets_update ON assets(update_id);
CREATE INDEX idx_assets_file_hash ON assets(file_hash);
CREATE INDEX idx_download_events_project ON download_events(project_id, timestamp DESC);
CREATE INDEX idx_download_events_update ON download_events(update_id);
CREATE INDEX idx_api_keys_lookup ON api_keys(key_suffix) INCLUDE (project_id, key_hash);