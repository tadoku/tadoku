-- Test safety: the seed must not leave its transaction bound to all tenants.
select set_config('tadoku.all_tenants', 'on', true);
