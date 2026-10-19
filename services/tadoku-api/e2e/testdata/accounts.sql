insert into tenants (key, kind)
values
  ('tadoku/identity-a-0123abcd', 'test'),
  ('tadoku/identity-b-0123abcd', 'test')
on conflict (key) do nothing;

with actors (id, display_name, created_at) as (
  values
    ('11111111-1111-4111-8111-111111111111'::uuid, 'Reader One', '2025-01-01 00:00:00Z'::timestamptz),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'Admin One', '2026-09-12 12:00:00Z'::timestamptz),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'Banned One', '2026-09-12 12:00:00Z'::timestamptz),
    ('44444444-4444-4444-8444-444444444444'::uuid, 'User Two', '2026-08-12 12:00:00Z'::timestamptz),
    ('55555555-5555-4555-8555-555555555555'::uuid, 'Young User', '2026-09-01 00:00:00Z'::timestamptz)
), scopes (tenant) as (
  values
    ('tadoku/prod'),
    ('tadoku/identity-a-0123abcd'),
    ('tadoku/identity-b-0123abcd')
)
insert into users (tenant, id, display_name, created_at, updated_at)
select scopes.tenant, actors.id, actors.display_name, actors.created_at, actors.created_at
from actors cross join scopes;

with actors (user_id, email) as (
  values
    ('11111111-1111-4111-8111-111111111111'::uuid, 'reader@example.test'),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'admin@example.test'),
    ('33333333-3333-4333-8333-333333333333'::uuid, 'banned@example.test'),
    ('44444444-4444-4444-8444-444444444444'::uuid, 'user2@example.test'),
    ('55555555-5555-4555-8555-555555555555'::uuid, 'young@example.test')
)
insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
select u.tenant, u.id, actors.email, actors.email,
  '$2b$12$TP8U/FQ8vCC/fzTCu5Fgve4WAHMgvG1jZ7V7bKWjDJRA.KeeC1xqK', u.created_at, u.updated_at
from users u
join actors on actors.user_id = u.id
where u.tenant in ('tadoku/prod', 'tadoku/identity-a-0123abcd', 'tadoku/identity-b-0123abcd');
