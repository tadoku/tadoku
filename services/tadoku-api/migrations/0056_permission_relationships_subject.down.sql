-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently permission_relationships_subject;
