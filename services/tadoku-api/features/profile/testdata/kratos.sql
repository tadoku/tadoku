with recursive identity_numbers(value) as (
  values (1)
  union all
  select value + 1 from identity_numbers where value < 501
)
insert into identities (id, nid, schema_id, traits, state, created_at, updated_at, state_changed_at)
select printf('10000000-0000-4000-8000-%012d', value), networks.id, 'user',
       printf('{"email":"user%03d@example.test","display_name":"User %03d"}', value, value), 'active',
       '2026-09-12 13:14:15+02:00', '2026-09-12 13:14:15+02:00', '2026-09-12 13:14:15+02:00'
from identity_numbers cross join networks;
