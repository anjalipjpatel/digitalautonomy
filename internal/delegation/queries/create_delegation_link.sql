INSERT INTO delegation_links (
    primary_user_id,
    delegate_user_id,
    valid_until,
    status,
    permissions
)
VALUES (
    @primary_user_id,
    @delegate_user_id,
    @valid_until,
    @status,
    @permissions
)
RETURNING
    id::text,
    primary_user_id::text,
    delegate_user_id::text,
    valid_from,
    valid_until,
    status::text,
    permissions;
