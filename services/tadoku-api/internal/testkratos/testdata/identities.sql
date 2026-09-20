insert into identities (id, nid, schema_id, traits, state, created_at, updated_at, state_changed_at)
select '11111111-1111-4111-8111-111111111111', id, 'user',
       '{"email":"reader@example.test","display_name":"Reader One"}', 'active',
       '2026-09-12 12:00:00', '2026-09-12 12:00:00', '2026-09-12 12:00:00'
from networks;
