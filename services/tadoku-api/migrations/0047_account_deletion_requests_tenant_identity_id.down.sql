-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently account_deletion_requests_tenant_identity_id;
