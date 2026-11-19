SELECT id::text, email, name, role::text, created_at
FROM users
WHERE id = @id;