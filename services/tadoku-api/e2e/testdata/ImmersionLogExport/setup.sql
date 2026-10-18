insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader', '2026-01-01', '2026-01-01'),
  ('tadoku/prod', '44444444-4444-4444-8444-444444444444', 'Other', '2026-01-01', '2026-01-01'),
  ('tadoku/prod', '22222222-2222-4222-8222-222222222222', 'Admin', '2026-01-01', '2026-01-01');
insert into logs (tenant, id, user_id, language_code, log_activity_id, amount, unit_key, unit_id, modifier, duration_seconds, computed_score, description, eligible_official_leaderboard, created_at, updated_at, deleted_at) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 10.5, 'reading_page', (select id from log_units where unit_key = 'reading_page' limit 1), 2, null, null, E'Comma, "quote"\n日本語', false, '2026-09-01T01:02:03.123456Z', '2026-09-01T01:02:03.123456Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 60, 0, null, false, '2026-09-01T01:02:03.123456Z', '2026-09-01T01:02:03.123456Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 2.5, 'reading_page', (select id from log_units where unit_key = 'reading_page' limit 1), 2, 90, 7.25, '=SUM(A1)', false, '2026-09-02T00:00:00Z', '2026-09-02T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 1, 0.5, '+text', false, '2026-09-03T00:00:00Z', '2026-09-03T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000005', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 30, 2, '-text', false, '2026-09-04T00:00:00Z', '2026-09-04T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000006', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 30, 2, '@text', false, '2026-09-05T00:00:00Z', '2026-09-05T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000007', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 30, 2, E'\ttext', false, '2026-09-06T00:00:00Z', '2026-09-06T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000008', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 30, 2, E'\rtext', false, '2026-09-07T00:00:00Z', '2026-09-07T00:00:00Z', null),
  ('tadoku/prod', 'a0000000-0000-4000-8000-000000000001', '22222222-2222-4222-8222-222222222222', 'eng', 2, null, null, null, null, 60, 3, 'Admin log', false, '2026-09-01T00:00:00Z', '2026-09-01T00:00:00Z', null),
  ('tadoku/prod', 'b0000000-0000-4000-8000-000000000001', '44444444-4444-4444-8444-444444444444', 'eng', 2, null, null, null, null, 60, 3, 'Other log', false, '2026-09-01T00:00:00Z', '2026-09-01T00:00:00Z', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000009', '11111111-1111-4111-8111-111111111111', 'eng', 2, null, null, null, null, 1, 0.5, '+text', false, '2026-09-03T00:00:00Z', '2026-09-03T00:00:00Z', '2026-09-08');
insert into log_tags (tenant, log_id, user_id, tag) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'zeta'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'comma, tag'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'alpha'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '=formula'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', '@tag');
insert into contests (tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end, registration_end, title, activity_type_id_allow_list, official) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Reader', false, '2026-09-01', '2026-09-30', '2026-09-30', 'Export fixture', '{1}', false);
insert into contest_logs (tenant, contest_id, log_id, duration_seconds, computed_score) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000001', 120, 999);
