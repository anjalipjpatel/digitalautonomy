SELECT
    id::text,
    primary_user_id::text,
    delegate_user_id::text,
    valid_from,
    valid_until,
    status::text,
    permissions
FROM delegation_links
WHERE delegate_user_id = @delegate_user_id
  AND status = 'pending'
  AND (valid_until IS NULL OR valid_until > CURRENT_TIMESTAMP)
ORDER BY valid_from;
