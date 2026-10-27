select
  accounts.email_normalized,
  accounts.user_id = '11111111-1111-4111-8111-111111111111' as reuses_deleted_user,
  users.display_name,
  users.deleted_at
from accounts
inner join users on users.tenant = accounts.tenant and users.id = accounts.user_id
where accounts.tenant = 'tadoku/prod'
  and accounts.email_normalized = 'reader@example.test'
order by accounts.user_id
