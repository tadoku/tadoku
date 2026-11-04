-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently posts_content_tenant_id;
