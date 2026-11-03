with fixture_accounts as (
  select attempts.number, gen_random_uuid() as id
  from generate_series(1, 49) as attempts(number)
), fixture_users as (
  insert into users (id, display_name, created_at, updated_at)
  select id, 'Throttle Fixture', '2026-01-01', '2026-01-01'
  from fixture_accounts
  returning id
)
insert into accounts (user_id, email, email_normalized, password_hash, created_at, updated_at)
select fixture.id,
  'unknown' || fixture.number || '@example.test',
  'unknown' || fixture.number || '@example.test',
  '$2b$04$Kr8WFdLSHO4NBQGmsmwqMe.2VlUTBQIIIXxP6HTw2nPq4IhuQ/P1W',
  '2026-01-01', '2026-01-01'
from fixture_accounts as fixture
join fixture_users as users on users.id = fixture.id;
