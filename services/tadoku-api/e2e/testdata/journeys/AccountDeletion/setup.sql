-- The deleted user is Reader One (11111111-…), signed in with the session below. User Two (44444444-…) is the
-- bystander. The users row matches the account cast, which runs after this file, so the session can reference it.
insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2025-01-01 00:00:00Z', '2025-01-01 00:00:00Z'),
  ('tadoku/prod', '99999999-9999-4999-8999-999999999999', 'Other Owner', '2026-01-01', '2026-01-01');

-- Test safety: the cookie value is tadoku_session=UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI, a synthetic token.
insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at, created_at) values
  ('tadoku/prod', '4458ef18ed1e6d9b4aca3df642ae4e4a2ccac196ff001e453b60c5cf2859c519',
   '11111111-1111-4111-8111-111111111111', '2026-09-12 11:30:00Z', '2026-10-12 11:30:00Z', '2026-09-12 11:30:00Z');

insert into contests (
  tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a1', '99999999-9999-4999-8999-999999999999', 'Other Owner', false,
   '2026-08-01', '2026-08-31', '2026-08-31', 'Finished official', '{}', '{1}', true,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a2', '99999999-9999-4999-8999-999999999999', 'Other Owner', false,
   '2026-08-25', '2026-09-30', '2026-09-30', 'Running', '{eng}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a3', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-30', 'Organized by the deleted user', '{eng}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a4', '99999999-9999-4999-8999-999999999999', 'Other Owner', false,
   '2025-12-01', '2025-12-31', '2025-12-31', 'Last year official', '{}', '{1}', true,
   '2025-01-01', '2025-01-01');

insert into contest_registrations (tenant, id, contest_id, user_id, language_codes) values
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b1', 'f0000000-0000-4000-8000-0000000000a1',
   '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b2', 'f0000000-0000-4000-8000-0000000000a2',
   '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b3', 'f0000000-0000-4000-8000-0000000000a3',
   '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b4', 'f0000000-0000-4000-8000-0000000000a4',
   '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b5', 'f0000000-0000-4000-8000-0000000000a2',
   '44444444-4444-4444-8444-444444444444', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000b6', 'f0000000-0000-4000-8000-0000000000a3',
   '44444444-4444-4444-8444-444444444444', '{eng}');

insert into logs (
  tenant, id, user_id, language_code, log_activity_id, duration_seconds, computed_score,
  eligible_official_leaderboard, description, created_at, updated_at, deleted_at
) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   600, 10, true, 'Finished reading', '2026-08-15', '2026-08-15', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   300, 5, false, 'Running reading', '2026-09-05', '2026-09-05', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   900, 15, true, 'Both contests', '2026-08-28', '2026-08-28', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   120, 2, false, 'No contest', '2026-09-10', '2026-09-10', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000005', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   60, 1, true, 'Deleted earlier', '2026-08-20', '2026-08-20', '2026-08-22'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000006', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   1200, 20, true, null, '2025-12-10', '2025-12-10', null),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000011', '44444444-4444-4444-8444-444444444444', 'eng', 1,
   300, 5, false, 'Bystander reading', '2026-09-06', '2026-09-06', null);

insert into contest_logs (tenant, contest_id, log_id, duration_seconds, computed_score) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a1', 'd0000000-0000-4000-8000-000000000001', 600, 10),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a2', 'd0000000-0000-4000-8000-000000000002', 300, 5),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a1', 'd0000000-0000-4000-8000-000000000003', 900, 15),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a2', 'd0000000-0000-4000-8000-000000000003', 900, 15),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a1', 'd0000000-0000-4000-8000-000000000005', 60, 1),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a4', 'd0000000-0000-4000-8000-000000000006', 1200, 20),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a2', 'd0000000-0000-4000-8000-000000000011', 300, 5),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000a3', 'd0000000-0000-4000-8000-000000000011', 300, 5);

insert into log_tags (tenant, log_id, user_id, tag) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'alpha'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'beta'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'gamma'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'delta'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-000000000011', '44444444-4444-4444-8444-444444444444', 'bystander');

insert into emails (tenant, id, template, recipient, subject, body, params, status, created_at) values
  ('tadoku/prod', 'c0000000-0000-4000-8000-000000000001', 'account.recovery', 'reader@example.test',
   'Recover your Tadoku account', 'Earlier message', '{"recovery_url":"https://account.tadoku.app/x"}', 'held',
   '2026-09-11 12:00:00Z');
