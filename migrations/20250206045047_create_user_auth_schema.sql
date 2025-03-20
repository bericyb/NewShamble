-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL CHECK (email ~* '^.+@.+\..+$'),
    full_name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Authentication table: Stores passwords and SSO provider info
CREATE TABLE user_auth (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    auth_provider TEXT NOT NULL DEFAULT 'local', -- 'local' for password, otherwise SSO provider name
    hashed_password TEXT, -- Nullable for SSO users
    provider_id TEXT UNIQUE, -- Used for SSO users (e.g., Google, GitHub IDs)
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CHECK (
        (auth_provider = 'local' AND hashed_password IS NOT NULL AND provider_id IS NULL) OR
        (auth_provider != 'local' AND hashed_password IS NULL AND provider_id IS NOT NULL)
    )
);

-- Sessions table: Tracks user logins
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL, -- Store JWT or session token
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Roles table: Basic role-based access control
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'user', 'moderator', 'guest')),
    assigned_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, role)
);
