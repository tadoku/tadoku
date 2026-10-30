-- name: DeleteUnfinishedRegistrationsForUser :many
delete from contest_registrations
using contests
where contests.id = contest_registrations.contest_id
  and contest_registrations.user_id = sqlc.arg(user_id)
  and contests.contest_end >= sqlc.arg(open_if_end_on_or_after)
returning contest_registrations.contest_id;

-- name: CountRegistrationsForUser :one
select count(*) from contest_registrations where user_id = sqlc.arg(user_id);

-- name: RenameContestsOwnedByUser :execrows
update contests
set owner_user_display_name = 'Deleted organizer'
where owner_user_id = sqlc.arg(user_id);
