

CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL,
    date_of_birth DATETIME DEFAULT NULL,
    avatar_url TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    about_me TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS followers (
    follower_id INTEGER NOT NULL,
    following_target_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_target_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_target_id) REFERENCES users(id) ON DELETE CASCADE
    -- CHECK (follower_id <> following_target_id)
);