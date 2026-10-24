insert into tenants (key, kind) values ('e2e/alpha-0000000a', 'test')
on conflict (key) do nothing;

insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2026-01-01', '2026-01-01'),
  ('e2e/alpha-0000000a', '11111111-1111-4111-8111-111111111111', 'Branch Reader', '2026-01-01', '2026-01-01');

insert into contests (
  tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest X', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'Reader One', true,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest Y', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Empty contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Deleted contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('e2e/alpha-0000000a', 'f1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Branch Reader', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Branch contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01');

update contests set deleted_at = '2026-09-10' where tenant = 'tadoku/prod' and id = 'f0000000-0000-4000-8000-000000000004';



insert into logs (
  tenant, id, user_id, language_code, log_activity_id, unit_id, unit_key, amount,
  modifier, computed_score, eligible_official_leaderboard, description, created_at, updated_at
) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   (select id from log_units where unit_key = 'reading_page' and language_code is null), 'reading_page', 10, 1, 20, false,
   'Contest preview', '2026-09-12', '2026-09-12'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   (select id from log_units where unit_key = 'reading_page' and language_code is null), 'reading_page', 3, 1, 6, false,
   'Unattached personal log', '2026-09-12', '2026-09-12'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   (select id from log_units where unit_key = 'reading_page' and language_code is null), 'reading_page', 4, 1, 8, false,
   'Deleted log', '2026-09-12', '2026-09-12'),
  ('e2e/alpha-0000000a', 'd1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   (select id from log_units where unit_key = 'reading_page' and language_code is null), 'reading_page', 7, 1, 14, false,
   'Branch-only preview', '2026-09-12', '2026-09-12');

update logs set deleted_at = '2026-09-12' where tenant = 'tadoku/prod' and id = 'd0000000-0000-4000-8000-000000000004';

insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, computed_score) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000001', 'reading_page', 10, 4, 40),
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000002', 'd0000000-0000-4000-8000-000000000001', 'reading_page', 10, 6, 60),
  ('tadoku/prod', 'f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000004', 'reading_page', 4, 4, 16),
  ('e2e/alpha-0000000a', 'f1000000-0000-4000-8000-000000000001', 'd1000000-0000-4000-8000-000000000001', 'reading_page', 7, 8, 56);

insert into log_tags (tenant, log_id, user_id, tag) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'book'),
  ('e2e/alpha-0000000a', 'd1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'branch');

insert into contest_registrations (tenant, id, contest_id, user_id, language_codes) values
  ('tadoku/prod', 'e0000000-0000-4000-8000-000000000001', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng,jpn}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-000000000002', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', '{eng,jpn}');
