-- Add unique constraints to email and username
ALTER TABLE users
ADD CONSTRAINT users_email_unique UNIQUE (email),
ADD CONSTRAINT users_username_unique UNIQUE (username);

---- create above / drop below ----

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_email_unique,
DROP CONSTRAINT IF EXISTS users_username_unique;