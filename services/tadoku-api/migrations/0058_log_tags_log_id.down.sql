-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently log_tags_log_id;
