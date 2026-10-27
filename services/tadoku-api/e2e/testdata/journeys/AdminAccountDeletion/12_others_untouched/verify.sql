select 'account' as kind, tenant, user_id::text as key,
  json_build_object(
    'email', email,
    'email_normalized', email_normalized,
    'password_hash', password_hash,
    'state', state,
    'created_at', created_at,
    'updated_at', updated_at
  ) as value
from accounts
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'contest', tenant, id::text,
  json_build_object(
    'owner_user_id', owner_user_id,
    'organizer', owner_user_display_name,
    'private', private,
    'contest_start', contest_start,
    'contest_end', contest_end,
    'registration_end', registration_end,
    'title', title,
    'language_code_allow_list', language_code_allow_list,
    'activity_type_id_allow_list', activity_type_id_allow_list,
    'official', official,
    'created_at', created_at,
    'updated_at', updated_at
  )
from contests
where not (tenant = 'tadoku/prod' and owner_user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'contest_log', contest_logs.tenant, contest_logs.contest_id::text || ' ' || contest_logs.log_id::text,
  json_build_object('duration_seconds', contest_logs.duration_seconds, 'score', contest_logs.computed_score)
from contest_logs
inner join logs on logs.tenant = contest_logs.tenant and logs.id = contest_logs.log_id
where not (logs.tenant = 'tadoku/prod' and logs.user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'log', tenant, id::text,
  json_build_object(
    'user_id', user_id,
    'language_code', language_code,
    'log_activity_id', log_activity_id,
    'duration_seconds', duration_seconds,
    'score', computed_score,
    'eligible_official_leaderboard', eligible_official_leaderboard,
    'description', description,
    'frozen_at', frozen_at,
    'created_at', created_at,
    'updated_at', updated_at,
    'deleted_at', deleted_at
  )
from logs
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'log_tag', tenant, log_id::text || ' ' || tag, json_build_object('user_id', user_id)
from log_tags
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'recovery_token', tenant, token_hash,
  json_build_object('user_id', user_id, 'expires_at', expires_at, 'used_at', used_at)
from recovery_tokens
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'registration', tenant, id::text,
  json_build_object('contest_id', contest_id, 'user_id', user_id, 'language_codes', language_codes)
from contest_registrations
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'relationship', tenant,
  object_type || ':' || object_id || '#' || relation || '@' || subject_type || ':' || subject_id,
  null
from permission_relationships
where not (
  tenant = 'tadoku/prod'
  and subject_type = 'user'
  and subject_id = '33333333-3333-4333-8333-333333333333'
)
union all
select 'session', tenant, token_hash,
  json_build_object(
    'user_id', user_id,
    'authenticated_at', authenticated_at,
    'expires_at', expires_at,
    'created_at', created_at
  )
from sessions
where not (tenant = 'tadoku/prod' and user_id = '33333333-3333-4333-8333-333333333333')
union all
select 'user', tenant, id::text,
  json_build_object(
    'display_name', display_name,
    'deletion_locked_at', deletion_locked_at,
    'deleted_at', deleted_at,
    'created_at', created_at,
    'updated_at', updated_at
  )
from users
where not (tenant = 'tadoku/prod' and id = '33333333-3333-4333-8333-333333333333')
order by kind, tenant, key
