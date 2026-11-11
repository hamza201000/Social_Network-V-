-- who sent the invitation (NULL for join requests)
ALTER TABLE members_status ADD COLUMN invited_by INTEGER DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL;
