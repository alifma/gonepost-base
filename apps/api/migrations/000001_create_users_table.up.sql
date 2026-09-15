CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email           text NOT NULL,
    username        text,
    full_name       text,
    password_hash   text NOT NULL,
    status          text NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Email is the canonical identifier: unique, case-insensitive.
CREATE UNIQUE INDEX users_email_key ON users (lower(email));

-- Username is optional but must be unique when set.
CREATE UNIQUE INDEX users_username_key ON users (lower(username)) WHERE username IS NOT NULL;
