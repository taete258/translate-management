-- Environments table
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Join table linking translation keys to environments
CREATE TABLE key_environments (
    key_id UUID NOT NULL REFERENCES translation_keys(id) ON DELETE CASCADE,
    env_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    PRIMARY KEY (key_id, env_id)
);

-- Indexes
CREATE INDEX idx_environments_project_id ON environments(project_id);
CREATE INDEX idx_key_environments_env_id ON key_environments(env_id);
CREATE INDEX idx_key_environments_key_id ON key_environments(key_id);
