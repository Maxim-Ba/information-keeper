CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE artifact_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO artifact_types (name) VALUES
    ('text'),
    ('loginpassword'),
    ('bankcard'),
    ('binary');

CREATE TABLE artifacts (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type_id INTEGER NOT NULL REFERENCES artifact_types(id),
    meta_info TEXT,
    link TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expired_at TIMESTAMP WITH TIME ZONE
);

-- Создание индексов отдельными командами
CREATE INDEX idx_artifacts_owner_id ON artifacts(owner_id);
CREATE INDEX idx_artifacts_type_id ON artifacts(type_id);
CREATE INDEX idx_artifacts_expired_at ON artifacts(expired_at);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_login ON users(login);
CREATE INDEX idx_artifacts_created_at ON artifacts(created_at);

UPDATE artifact_types SET id = 0 WHERE name = 'text';
UPDATE artifact_types SET id = 1 WHERE name = 'loginpassword'; 
UPDATE artifact_types SET id = 2 WHERE name = 'bankcard';
UPDATE artifact_types SET id = 3 WHERE name = 'binary';
