-- Test safety: deliberately unscoped; the seeding role has no privileges on jobs, whose policy honours all tenants.
delete from jobs;
