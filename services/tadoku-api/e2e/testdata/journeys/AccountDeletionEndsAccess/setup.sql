-- The users row matches testdata/accounts.sql, which runs after this file, so the session and recovery token can
-- reference Reader One.
insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader One', '2025-01-01 00:00:00Z', '2025-01-01 00:00:00Z');

-- Test safety: the cookie value is tadoku_session=UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI, a synthetic token.
insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at, created_at) values
  ('tadoku/prod', '4458ef18ed1e6d9b4aca3df642ae4e4a2ccac196ff001e453b60c5cf2859c519',
   '11111111-1111-4111-8111-111111111111', '2026-09-12 11:30:00Z', '2026-10-12 11:30:00Z', '2026-09-12 11:30:00Z');

-- Test safety: the recovery path token is ZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGQ, a synthetic token.
insert into recovery_tokens (tenant, token_hash, user_id, expires_at) values
  ('tadoku/prod', 'dfd8f6e54f5fb5ab612b82073514426326e48e1ac62a8d62d57b1ffa48639d04',
   '11111111-1111-4111-8111-111111111111', '2026-09-12 13:00:00Z');
