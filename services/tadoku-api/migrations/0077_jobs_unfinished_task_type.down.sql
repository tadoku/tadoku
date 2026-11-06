-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently jobs_unfinished_task_type;
