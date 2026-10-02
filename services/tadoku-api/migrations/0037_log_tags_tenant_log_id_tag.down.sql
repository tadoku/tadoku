-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently log_tags_tenant_log_id_tag;
