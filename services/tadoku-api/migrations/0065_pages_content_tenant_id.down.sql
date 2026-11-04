-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently pages_content_tenant_id;
