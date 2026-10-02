-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently profiles_tenant_user_id;
