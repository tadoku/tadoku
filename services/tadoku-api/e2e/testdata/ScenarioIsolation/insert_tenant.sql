-- Test safety: the seeding role must refuse registering a tenant.
insert into tenants (key, kind) values ('e2e/x-0123abcd', 'test');
