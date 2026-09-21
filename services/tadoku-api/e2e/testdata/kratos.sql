-- Standard cast, matching the fixed subjects and traits in the signed JWTs.
-- Kratos initializes the only network before this seed runs. Its admin API
-- cannot choose identity UUIDs, so these test-only rows use its pinned schema.
with actors(id, email, display_name, created_at) as (
    values
        ('11111111-1111-4111-8111-111111111111', 'reader@example.test', 'Reader One', '2025-01-01 00:00:00'),
        ('22222222-2222-4222-8222-222222222222', 'admin@example.test', 'Admin One', '2026-09-12 12:00:00'),
        ('33333333-3333-4333-8333-333333333333', 'banned@example.test', 'Banned One', '2026-09-12 12:00:00'),
        ('44444444-4444-4444-8444-444444444444', 'user2@example.test', 'User Two', '2026-08-12 12:00:00'),
        ('55555555-5555-4555-8555-555555555555', 'young@example.test', 'Young User', '2026-09-01 00:00:00')
)
insert into identities (id, nid, schema_id, traits, state, created_at, updated_at, state_changed_at)
select actors.id, networks.id, 'user',
       json_object('email', actors.email, 'display_name', actors.display_name), 'active',
       actors.created_at, actors.created_at, actors.created_at
from actors cross join networks;
