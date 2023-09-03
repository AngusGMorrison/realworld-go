CREATE TABLE IF NOT EXISTS users (
     id TEXT NOT NULL PRIMARY KEY,
     username TEXT NOT NULL UNIQUE,
     email TEXT NOT NULL UNIQUE,
     password_hash TEXT NOT NULL,
     bio TEXT,
     image_url TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx
ON users (email);