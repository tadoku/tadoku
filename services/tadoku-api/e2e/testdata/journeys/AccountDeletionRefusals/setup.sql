-- The users rows match the account cast, which runs after this file, so the sessions can reference them.
insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2025-01-01 00:00:00Z', '2025-01-01 00:00:00Z'),
  ('tadoku/prod', '22222222-2222-4222-8222-222222222222', 'Admin One', '2026-09-12 12:00:00Z', '2026-09-12 12:00:00Z'),
  ('tadoku/prod', '33333333-3333-4333-8333-333333333333', 'Banned One', '2026-09-12 12:00:00Z', '2026-09-12 12:00:00Z');

-- Test safety: synthetic session tokens. Reader One holds a fresh session (cookie UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI)
-- and one authenticated an hour and a second ago (cookie U1NTU1NTU1NTU1NTU1NTU1NTU1NTU1NTU1NTU1NTU1M). Admin One uses
-- QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE and Banned One QkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkI.
insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at, created_at) values
  ('tadoku/prod', '4458ef18ed1e6d9b4aca3df642ae4e4a2ccac196ff001e453b60c5cf2859c519', '11111111-1111-4111-8111-111111111111',
   '2026-09-12 11:30:00Z', '2026-10-12 11:30:00Z', '2026-09-12 11:30:00Z'),
  ('tadoku/prod', '7e6ee326ce17dca627fe5a20ce8b7f35e6a0da0afd4cd22ec2dd802d84eb52d2', '11111111-1111-4111-8111-111111111111',
   '2026-09-12 10:59:59Z', '2026-10-12 10:59:59Z', '2026-09-12 10:59:59Z'),
  ('tadoku/prod', 'aebddc7466d56ad82b5797e7bb90a4224bfa2bd3c12d364d88fc2416854009dc', '22222222-2222-4222-8222-222222222222',
   '2026-09-12 11:30:00Z', '2026-10-12 11:30:00Z', '2026-09-12 11:30:00Z'),
  ('tadoku/prod', '76948896d1f54257aeae6ddfff136452705376ac2655569f36786e624b7ea0d1', '33333333-3333-4333-8333-333333333333',
   '2026-09-12 11:30:00Z', '2026-10-12 11:30:00Z', '2026-09-12 11:30:00Z');
