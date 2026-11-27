SELECT
    id::text,
    primary_user_id::text,
    delegate_user_id::text,
    valid_from,
    valid_until,
    status::text,
    permissions
FROM delegation_links
WHERE primary_user_id = @user_id
   OR delegate_user_id = @user_id
ORDER BY valid_from DESC;
