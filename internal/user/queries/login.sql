SELECT id::text, email, name, role::text, created_at, hashed_password
FROM users
WHERE email = @email;