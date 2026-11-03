-- Test safety: widens the binding to every tenant around an unscoped delete; scenario seeding must refuse it.
select set_config('tadoku.all_tenants', 'on', true);
delete from jobs;
select set_config('tadoku.all_tenants', 'off', true);
