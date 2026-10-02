-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently contest_logs_tenant_contest_id_log_id;
