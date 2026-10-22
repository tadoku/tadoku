insert into tenants (key, kind) values ('e2e/isolation-0123abcd', 'test') on conflict (key) do nothing;
