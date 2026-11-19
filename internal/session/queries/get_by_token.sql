SELECT user_id, expires_at, revoked
FROM sessions
WHERE token_hash = @token_hash;
