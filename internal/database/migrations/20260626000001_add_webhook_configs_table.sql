-- +goose Up
CREATE TABLE fluxend.webhook_configs (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID REFERENCES authentication.users(uuid),
    updated_by UUID REFERENCES authentication.users(uuid),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_webhook_configs_project FOREIGN KEY (project_uuid)
        REFERENCES fluxend.projects(uuid) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS fluxend.webhook_configs CASCADE;
