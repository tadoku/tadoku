insert into tenants (key, kind) values ('e2e/alpha-0000000a', 'test');
insert into tenant_overrides (tenant, component) values ('e2e/alpha-0000000a', 'tadoku-worker');

delete from platform_scoring_config;
delete from scoring_rules;
delete from scoring_rule_sets;

insert into scoring_rule_sets (
  tenant, id, scope, version, status, created_at, published_at
) values
  ('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001',
   'platform', 1, 'published', '2026-01-01', '2026-01-02'),
  ('e2e/alpha-0000000a', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000002',
   'platform', 1, 'published', '2026-01-01', '2026-01-02');

insert into scoring_rules (
  tenant, id, rule_set_id, priority, stackable, activity_id, unit_key, score_source, rate
) values
  ('tadoku/prod', 'bbbbbbbb-bbbb-4bbb-8bbb-000000000001', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001',
   1, false, 1, 'reading_page', 'amount', 2),
  ('e2e/alpha-0000000a', 'bbbbbbbb-bbbb-4bbb-8bbb-000000000002', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000002',
   1, false, 1, 'reading_page', 'amount', 2);

insert into platform_scoring_config (tenant, singleton, active_rule_set_id) values
  ('tadoku/prod', true, 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001'),
  ('e2e/alpha-0000000a', true, 'aaaaaaaa-aaaa-4aaa-8aaa-000000000002');

insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2026-01-01', '2026-01-01'),
  ('e2e/alpha-0000000a', '44444444-4444-4444-8444-444444444444', 'User Two', '2026-01-01', '2026-01-01');

insert into contests (
  tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001',
   '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Production contest', '{eng}', '{1}',
   true, '2026-01-01', '2026-01-01'),
  ('e2e/alpha-0000000a', 'f0000000-0000-4000-8000-000000000002',
   '44444444-4444-4444-8444-444444444444', 'User Two', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Test contest', '{eng}', '{1}',
   true, '2026-01-01', '2026-01-01');

insert into contest_registrations (tenant, id, contest_id, user_id, language_codes) values
  ('tadoku/prod', 'eeeeeeee-eeee-4eee-8eee-000000000001',
   'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('e2e/alpha-0000000a', 'eeeeeeee-eeee-4eee-8eee-000000000002',
   'f0000000-0000-4000-8000-000000000002', '44444444-4444-4444-8444-444444444444', '{eng}');

update log_units
set id = '026e644b-ec51-61cd-a392-55231a6ab4a0'
where unit_key = 'reading_page' and language_code is null;
