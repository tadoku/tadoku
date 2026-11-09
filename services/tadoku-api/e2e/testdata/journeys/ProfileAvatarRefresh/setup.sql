-- The cast's users rows come first so that their accounts can carry the Gravatar fixture emails.
insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Reader One', '2025-01-01 00:00:00Z', '2025-01-01 00:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', 'Admin One', '2026-09-12 12:00:00Z', '2026-09-12 12:00:00Z'),
  ('44444444-4444-4444-8444-444444444444', 'User Two', '2026-08-12 12:00:00Z', '2026-08-12 12:00:00Z');

insert into accounts (user_id, email, email_normalized, password_hash, created_at, updated_at)
select u.id, emails.email, emails.email,
  '$2b$04$Kr8WFdLSHO4NBQGmsmwqMe.2VlUTBQIIIXxP6HTw2nPq4IhuQ/P1W', u.created_at, u.updated_at
from users u
join (
  values
    ('11111111-1111-4111-8111-111111111111'::uuid, 'journey-avatar@example.test'),
    ('22222222-2222-4222-8222-222222222222'::uuid, 'journey-error@example.test'),
    ('44444444-4444-4444-8444-444444444444'::uuid, 'journey-none@example.test')
) as emails (user_id, email) on emails.user_id = u.id;
