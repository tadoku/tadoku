insert into tenants (key, kind) values ('e2e/flipt-0123abcd', 'test') on conflict (key) do nothing;
