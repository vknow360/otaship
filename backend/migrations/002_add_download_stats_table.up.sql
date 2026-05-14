CREATE TABLE download_stats (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    update_id UUID NOT NULL REFERENCES updates(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,
    channel TEXT NOT NULL,
    date DATE NOT NULL,
    download_count INT NOT NULL DEFAULT 0,
    UNIQUE (project_id, update_id, platform, channel, date)
);

CREATE INDEX idx_download_stats_project ON download_stats(project_id);