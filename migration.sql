-- Migration script for Admin Dashboard & RBAC

-- Pre-existing tables (predate this migration file in production; included
-- here with IF NOT EXISTS so this script can also bootstrap a fresh database,
-- e.g. for local Docker development).
CREATE TABLE IF NOT EXISTS usernames (
    uuid VARCHAR(100) PRIMARY KEY,
    username TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    post_id INT NOT NULL,
    comment TEXT NOT NULL,
    post_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create posts table
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    blurb TEXT NOT NULL,
    content TEXT NOT NULL,
    date_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Fix pre-existing databases created before this migration used TIMESTAMPTZ
-- (postgresql-simple only decodes TIMESTAMPTZ columns into Haskell's UTCTime;
-- a plain TIMESTAMP column throws a decode error at query time).
ALTER TABLE posts ALTER COLUMN date_time TYPE TIMESTAMPTZ USING date_time AT TIME ZONE 'UTC';
ALTER TABLE comments ALTER COLUMN post_time TYPE TIMESTAMPTZ USING post_time AT TIME ZONE 'UTC';

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

-- Create user_roles table
CREATE TABLE IF NOT EXISTS user_roles (
    user_uuid VARCHAR(100) NOT NULL,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_uuid, role_id)
);

-- Insert default roles if they don't exist
INSERT INTO roles (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('moderator') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('writer') ON CONFLICT (name) DO NOTHING;

-- Track which user authored each post
ALTER TABLE posts ADD COLUMN IF NOT EXISTS author_uuid VARCHAR(100);

-- Grant the default site owner admin + writer on any fresh/migrated database
INSERT INTO user_roles (user_uuid, role_id)
SELECT 'G3QzWiV57VgYwAKbMTa07OaX3ra2', id FROM roles WHERE name IN ('admin', 'writer')
ON CONFLICT DO NOTHING;
