delete from platform_scoring_config;
delete from scoring_rules;
delete from scoring_rule_sets;

insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Reader One', '2026-01-01', '2026-01-01'),
  ('44444444-4444-4444-8444-444444444444', 'Reader Two', '2026-01-01', '2026-01-01');

insert into scoring_rule_sets (id, scope, version, status, created_at, published_at) values
  ('aaaaaaaa-aaaa-4aaa-8aaa-000000000001', 'platform', 1, 'published',
   '2026-01-01', '2026-01-02');

insert into scoring_rules (
  id, rule_set_id, priority, stackable, activity_id, unit_key, score_source, rate
) values
  ('bbbbbbbb-bbbb-4bbb-8bbb-000000000001', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001',
   1, false, 1, 'pages', 'amount', 2);

insert into platform_scoring_config (singleton, active_rule_set_id) values
  (true, 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001');

insert into contests (
  id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest X', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'Reader One', true,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest Y', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01');

insert into scoring_rule_sets (
  id, scope, contest_id, version, status, mode, created_at, published_at
) values
  ('cccccccc-cccc-4ccc-8ccc-000000000001', 'contest', 'f0000000-0000-4000-8000-000000000001',
   1, 'published', 'replace', '2026-01-01', '2026-01-02');

insert into scoring_rules (
  id, rule_set_id, priority, stackable, activity_id, unit_key, score_source, rate
) values
  ('dddddddd-dddd-4ddd-8ddd-000000000001', 'cccccccc-cccc-4ccc-8ccc-000000000001',
   1, false, 1, 'pages', 'amount', 4);

update contests set scoring_rule_set_id = 'cccccccc-cccc-4ccc-8ccc-000000000001'
where id = 'f0000000-0000-4000-8000-000000000001';

insert into contest_registrations (id, contest_id, user_id, language_codes) values
  ('e0000000-0000-4000-8000-000000000001', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng,jpn}'),
  ('e0000000-0000-4000-8000-000000000002', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', '{eng,jpn}');

insert into logs (
  id, user_id, language_code, log_activity_id, unit_key, amount, modifier,
  computed_score, eligible_official_leaderboard, description, created_at, updated_at
) values
  ('d0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   'pages', 10, 1, 20, false,
   'unchanged', '2026-09-12', '2026-09-12'),
  ('d0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'jpn', 1,
   'pages', 5, 1, 10, false,
   'Japanese log', '2026-09-12', '2026-09-12'),
  ('d0000000-0000-4000-8000-000000000003', '44444444-4444-4444-8444-444444444444', 'eng', 1,
   'pages', 3, 1, 6, false,
   'Other user log', '2026-09-12', '2026-09-12');

insert into contest_logs (contest_id, log_id, unit_key, amount, modifier, computed_score) values
  ('f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000001', 'pages', 10, 1, 40),
  ('f0000000-0000-4000-8000-000000000002', 'd0000000-0000-4000-8000-000000000001', 'pages', 10, 1, 20),
  ('f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000002', 'pages', 5, 1, 20),
  ('f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000003', 'pages', 3, 1, 12);

insert into log_tags (log_id, user_id, tag) values
  ('d0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'book');
