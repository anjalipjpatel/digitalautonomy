UPDATE sessions
SET revoked = TRUE
WHERE token_hash = @token_hash;