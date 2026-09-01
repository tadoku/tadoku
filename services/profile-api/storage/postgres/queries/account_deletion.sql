-- name: DeleteProfileForAccountDeletion :exec
delete from profiles
where user_id = sqlc.arg('identity_id');

-- name: ListAccountDeletionSuppressedIdentityIDs :many
select identity_id
from account_deletion_requests
where status in (
  'queued',
  'access_locked',
  'immersion_scrubbed',
  'caches_reconciled',
  'authorization_removed',
  'identity_deleted',
  'complete',
  'manual_attention'
)
order by identity_id;
