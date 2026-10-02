-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently contest_registrations_tenant_user_id_contest_id;
