-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently users_tenant_id;
