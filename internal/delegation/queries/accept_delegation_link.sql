UPDATE delegation_links
SET status = 'active'
WHERE id = @id
  AND delegate_user_id = @delegate_user_id
  AND status = 'pending'
  AND (valid_until IS NULL OR valid_until > CURRENT_TIMESTAMP)
RETURNING
    id::text,
    primary_user_id::text,
    delegate_user_id::text,
    valid_from,
    valid_until,
    status::text,
    permissions;
