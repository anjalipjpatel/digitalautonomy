INSERT INTO users (email, hashed_password, name, role)
VALUES (@email, @hashed_password, @name, @role)
RETURNING id::text, email, name, role::text, created_at;