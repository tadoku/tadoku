-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently moderation_audit_log_tenant_id;
