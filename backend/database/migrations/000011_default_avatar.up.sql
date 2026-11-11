-- users without a profile picture get the default icon
UPDATE users SET avatar_url = 'icon_user.png' WHERE avatar_url = '' OR avatar_url IS NULL;
