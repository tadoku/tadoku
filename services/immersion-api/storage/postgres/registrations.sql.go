// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: registrations.sql

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const findContestRegistrationForUser = `-- name: FindContestRegistrationForUser :one
select
  id,
  contest_id,
  user_id,
  user_display_name,
  language_codes
from contest_registrations
where
  user_id = $1
  and contest_id = $2
  and deleted_at is null
`

type FindContestRegistrationForUserParams struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type FindContestRegistrationForUserRow struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
}

func (q *Queries) FindContestRegistrationForUser(ctx context.Context, arg FindContestRegistrationForUserParams) (FindContestRegistrationForUserRow, error) {
	row := q.db.QueryRowContext(ctx, findContestRegistrationForUser, arg.UserID, arg.ContestID)
	var i FindContestRegistrationForUserRow
	err := row.Scan(
		&i.ID,
		&i.ContestID,
		&i.UserID,
		&i.UserDisplayName,
		pq.Array(&i.LanguageCodes),
	)
	return i, err
}

const findOngoingContestRegistrationForUser = `-- name: FindOngoingContestRegistrationForUser :many
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.user_display_name,
  contest_registrations.language_codes,
  contests.activity_type_id_allow_list,
  contests.registration_end,
  contests.contest_start,
  contests.contest_end,
  contests.private,
  contests.title,
  contests.description
from contest_registrations
inner join contests
  on contests.id = contest_registrations.contest_id
where
  user_id = $1
  and contests.contest_start <= $2::timestamp
  and (contests.contest_end + '1 day'::interval) > $2::timestamp
  and contest_registrations.deleted_at is null
`

type FindOngoingContestRegistrationForUserParams struct {
	UserID uuid.UUID
	Now    time.Time
}

type FindOngoingContestRegistrationForUserRow struct {
	ID                      uuid.UUID
	ContestID               uuid.UUID
	UserID                  uuid.UUID
	UserDisplayName         string
	LanguageCodes           []string
	ActivityTypeIDAllowList []int32
	RegistrationEnd         time.Time
	ContestStart            time.Time
	ContestEnd              time.Time
	Private                 bool
	Title                   string
	Description             sql.NullString
}

func (q *Queries) FindOngoingContestRegistrationForUser(ctx context.Context, arg FindOngoingContestRegistrationForUserParams) ([]FindOngoingContestRegistrationForUserRow, error) {
	rows, err := q.db.QueryContext(ctx, findOngoingContestRegistrationForUser, arg.UserID, arg.Now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindOngoingContestRegistrationForUserRow
	for rows.Next() {
		var i FindOngoingContestRegistrationForUserRow
		if err := rows.Scan(
			&i.ID,
			&i.ContestID,
			&i.UserID,
			&i.UserDisplayName,
			pq.Array(&i.LanguageCodes),
			pq.Array(&i.ActivityTypeIDAllowList),
			&i.RegistrationEnd,
			&i.ContestStart,
			&i.ContestEnd,
			&i.Private,
			&i.Title,
			&i.Description,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertContestRegistration = `-- name: UpsertContestRegistration :one
insert into contest_registrations (
  id,
  contest_id,
  user_id,
  user_display_name,
  language_codes
) values (
  $1,
  $2,
  $3,
  $4,
  $5
) on conflict (id) do
update set
  language_codes = $5,
  updated_at = now()
returning id
`

type UpsertContestRegistrationParams struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
}

func (q *Queries) UpsertContestRegistration(ctx context.Context, arg UpsertContestRegistrationParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, upsertContestRegistration,
		arg.ID,
		arg.ContestID,
		arg.UserID,
		arg.UserDisplayName,
		pq.Array(arg.LanguageCodes),
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}
