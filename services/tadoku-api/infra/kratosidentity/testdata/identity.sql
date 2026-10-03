insert into identities (id, nid, schema_id, traits, state, metadata_admin, metadata_public,
                        created_at, updated_at, state_changed_at)
select '11111111-1111-4111-8111-111111111111', id, 'user',
       '{"email":"reader@example.test","display_name":"Reader One"}', 'active',
       '{"support":"keep"}', '{"theme":"keep"}',
       '2026-09-12 12:00:00', '2026-09-12 12:00:00', '2026-09-12 12:00:00'
from networks;

insert into sessions (id, nid, identity_id, issued_at, expires_at, authenticated_at,
                      created_at, updated_at, active, token, logout_token, aal, authentication_methods)
select '22222222-2222-4222-8222-222222222222', id, '11111111-1111-4111-8111-111111111111',
       '2026-09-12 12:00:00', '2099-09-12 12:00:00', '2026-09-12 12:00:00',
       '2026-09-12 12:00:00', '2026-09-12 12:00:00', true,
       'disposable-kratos-writer-session-token', 'disposable-kratos-writer-logout-token',
       'aal1', '[{"method":"password","aal":"aal1","completed_at":"2026-09-12T12:00:00Z"}]'
from networks;
