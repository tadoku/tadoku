-- name: LockTenant :exec
select pg_advisory_xact_lock(hashtextextended('tadoku-tenant:' || sqlc.arg('key')::text, 0));

-- name: OwnsTenantRegistry :one
select pg_has_role(current_user, relowner, 'USAGE')::boolean as owns_registry
from pg_catalog.pg_class where oid = 'tenants'::regclass;

-- name: TenantKind :one
select kind from tenants where key = $1;

-- name: InsertTestTenant :exec
insert into tenants (key, kind) values ($1, 'test') on conflict do nothing;

-- name: DeleteTestTenant :exec
delete from tenants where key = $1 and kind = 'test';

-- name: SetOverride :exec
insert into tenant_overrides (tenant, component) values ($1, $2) on conflict do nothing;

-- name: ClearOverride :exec
delete from tenant_overrides where tenant = $1 and component = $2;
