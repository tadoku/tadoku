-- Test safety: deliberately unscoped; scenario seeding must confine it to the seeding tenant.
delete from sessions;
delete from accounts;
delete from permission_relationships;
